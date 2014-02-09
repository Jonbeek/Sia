package disk

type DiskStorage interface{
	CreateFile(filehash string, length uint64) error
	ReadFile(filehash string, offset uint64, length uint64) []byte
	WriteFile(filehash string, start uint64, data []byte) error
}




















