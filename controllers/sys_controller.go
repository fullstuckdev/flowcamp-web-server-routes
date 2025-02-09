package controllers

import (
	"fmt"
	"golangapi/models"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"io"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type SysController struct {
	DB *gorm.DB
	downloadMutex sync.Mutex
}

func NewSysController(db *gorm.DB) *SysController {
	return &SysController{DB: db}
}

func (sc *SysController) CreateDirectory(c *gin.Context) {
	var req models.CreateDirectoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Nama folder tidak boleh kosong"})
		return
	}

	err := os.Mkdir(req.DirectoryName, 0755) 

	if err != nil {
		c.JSON(400, gin.H{"error": "Gagal membuat directory!"})
		return
	}

	c.JSON(201, gin.H{"message": "Folder berhasil dibuat"})

}

func (sc *SysController) CreateFile(c *gin.Context) {
	var req models.CreateFileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "semua field wajib di isi.."})
		return
	}

	if err := os.MkdirAll(req.DirectoryName, 0755); err != nil {
		c.JSON(400, gin.H{"error": "gagal membuat folder"})
		return
	}

	filepath := filepath.Join(req.DirectoryName, req.FileName)

	file, err := os.Create(filepath)
	if err != nil {
		c.JSON(400, gin.H{"error": "Gagal membuat file"})
		return
	}

	defer file.Close()

	_, err = file.WriteString(req.Content)
	if err != nil {
		c.JSON(400, gin.H{"error": "Gagal mmenulis file"})
		return
	}

	c.JSON(201, gin.H{
	"message": "File berhasil dibuat dan ditulis", 
	"path": filepath,
	})
}

func (sc *SysController) ReadFile(c *gin.Context) {
	var req models.ReadFileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Nama folder dan nama file wajib di isi.."})
		return
	}

	filePath := filepath.Join(req.DirectoryName, req.FileName)

	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		c.JSON(400, gin.H{"error": "Error membaca data"})
		return
	}

	c.JSON(200, gin.H{"data": string(data)})
}

func (sc *SysController) RenameFile(c *gin.Context) {
	var req models.RenameFileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "directory name, old name, dan new name harus terisi!"})
		return
	}

	oldPath := filepath.Join(req.DirectoryName, req.OldFileName)
	newPath := filepath.Join(req.DirectoryName, req.NewFileName)

	err := os.Rename(oldPath, newPath)

	if err != nil {
		c.JSON(400, gin.H{"error": "Error mengubah nama file dan folder"})
		return
	}

	c.JSON(200, gin.H{
		"message": "data berhasil dirubah",
		"old_path": oldPath,
		"new_path": newPath,
	})
}

func (sc *SysController) UploadFile(c *gin.Context) {
	const uploaadDir = "uploads"
	if err := os.MkdirAll(uploaadDir, 0755); err != nil {
		c.JSON(400, gin.H{"error": "Gagal membuat directory"})
		return
	}

	file, err := c.FormFile("file")

	if err != nil {
		c.JSON(400, gin.H{"error": "File tidak boleh kosong"})
		return
	}

	filename := filepath.Base(file.Filename)

	//"example_dir3/belajarGolang-Edit.txt", (contoh)
	PathFolder := filepath.Join(uploaadDir, filename)

	if err := c.SaveUploadedFile(file, PathFolder); err != nil {
		c.JSON(400, gin.H{"error": "Error mengupload file"})
		return
	}

	response := models.FileUploadResponse{
		Message: "File berhasil diupload!",
		Filename: filename,
		Path: PathFolder,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})

}

func (sc *SysController) DownloadFile(c *gin.Context) {

	// check userId

	// check email

	// if email == "taufikmul@gmail.com"

	// dia bisa akses file apa aja sama folder apa saja

	// else

	// dia tidak bisa akses folder dan file

	sc.downloadMutex.Lock()
	defer sc.downloadMutex.Unlock()

	fileName := c.Query("file_name")
	dirName := c.Query("directory_name")

	if fileName == "" || dirName == "" {
		c.JSON(400, gin.H{"error": "Filename dan directory name tidak boleh kosong!"})
		return
	}

	//"example_dir3/belajarGolang-Edit.txt", (contoh)
	filePath := filepath.Join(dirName, fileName)

	absPath, err := filepath.Abs(filePath)

	if err != nil {
		c.JSON(400, gin.H{"error": "Folder tidak ditemukan!"})
		return
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		c.JSON(400, gin.H{"error": "File tidak ditemukan!"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Type", "application,octet-stream")

	// bikin channel
	done := make(chan bool)
	errChan := make(chan error)

	// bikin fungsi async
	go func() {
		file, err := os.Open(absPath)

		if err != nil {
			c.JSON(400, gin.H{"error": "Gagal membuka file!"})
			return
		}

		defer file.Close()

		fileInfo, err := file.Stat()

		if err != nil {
			c.JSON(400, gin.H{"error": "Gagal membaca file"})
			return
		}

		c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

		buffer := make([]byte, 32 * 1024)
		for{
			n, err := file.Read(buffer)
			if err == io.EOF {
				done <- true
				return
			}

			if err != nil {
				errChan <- fmt.Errorf("error reading file")
				return
			}

			if _, err := c.Writer.Write(buffer[:n]); err != nil {
				errChan <- fmt.Errorf("error writing to response")
				return
			}
			c.Writer.Flush()
		}
	}()

	select {
	case <- done:
		return
	case err := <- errChan:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	case <-time.After(5 * time.Minute):
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "Download timeout"})
		return
	}

}