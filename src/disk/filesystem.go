package disk

import (
	"os"
	"encoding/json"
)

type SwarmStorage struct {
	SwarmId    string
	amountused uint64
	files map[string]uint64
}

//helper function to produce the correct filename
func (r SwarmStorage) getFileName(filehash string) string {
	return r.SwarmId + string(os.PathSeparator) + filehash
}

//Opens or creates directory for swarm info, and if it exists, obtains the correct amount of space used by its
//files
func CreateSwarmSystem(swarmid string) (r *SwarmStorage, err error) {
	var i SwarmStorage
	i.SwarmId = swarmid
	i.amountused = 0
	err = os.Mkdir(swarmid, os.ModeDir|os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			o, e := os.Open(swarmid + string(os.PathSeparator))
			if e != nil {
			}
			defer o.Close()
			fileInfo, e := o.Readdir(-1)
			for _, fi := range fileInfo {
				if fi.Mode().IsRegular() {
					i.amountused += uint64(fi.Size())
				}
			}
		}
	}
	r = &i
	return
}

func (r SwarmStorage) CreateFile(filehash string, length uint64) (written int64, err error) {
	file, err := os.Create(r.SwarmId + string(os.PathSeparator) + filehash)

	if err != nil && os.IsExist(err) {
		//in which case, it should be safe to ignore the error
		err = nil
	}
	defer file.Close()
	err = file.Truncate(int64(length))
	r.files[filehash]=uint64(length)
	written = int64(length)
	return
}

func (r SwarmStorage) DeleteFile(filehash string) error {
	size, err := os.Stat(r.getFileName(filehash))
	if err == nil {
		r.amountused -= uint64(size.Size())
		err = os.Remove(r.SwarmId + string(os.PathSeparator) + filehash)
	}
	r.files[filehash]=uint64(0)
	return err
}

func (r SwarmStorage) WriteFile(filehash string, start uint64, data []byte) error {
	path := r.SwarmId + string(os.PathSeparator) + filehash
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	size,ok:=r.files[filehash]
	if uint64(start)+uint64(len(data))>=size&&ok{
		r.amountused+=uint64(start)+uint64(uint64(len(data))-size)
		r.files[filehash]=uint64(start)+uint64(len(data))
	}
	file.WriteAt(data, int64(start))
	file.Close()
	return nil

}
func (r SwarmStorage) ReadFile(filehash string, start uint64, data []byte) (err error) {
	file, err := os.Open(r.getFileName(filehash))
	file.ReadAt(data, int64(start))
	return
}
func (r SwarmStorage) SaveSwarm(){
	s,err:=os.Open(r.SwarmId+".conf")
	if err!=nil{
		//do something in the future.
	}
	defer s.Close()
	js:=json.NewEncoder(s)
	if err=js.Encode(&s);err!=nil{
		print(err.Error())
	}

}










