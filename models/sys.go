package models

type CreateDirectoryRequest struct {
	DirectoryName string `json:"directory_name" binding:"required"`
}

type CreateFileRequest struct {
	DirectoryName string `json:"directory_name" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type ReadFileRequest struct {
	DirectoryName string `json:"directory_name" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
}

type RenameFileRequest struct {
	DirectoryName string `json:"directory_name" binding:"required"`
	OldFileName string `json:"old_file_name" binding:"required"`
	NewFileName string `json:"new_file_name" binding:"required"`
}

type FileUploadResponse struct {
	Message string `json:"message"`
	Filename string `json:"filename"`
	Path string `json:"path"`
}

type DownloadFileRequest struct {
	DirectoryName string `json:"directory_name" binding:"required"`
	FileName string `json:"file_name" binding:"required"`
}