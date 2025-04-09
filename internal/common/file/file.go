package file

// Another its name is file type
type FileExtension string

func (f FileExtension) String() string {
	return string(f)
}
