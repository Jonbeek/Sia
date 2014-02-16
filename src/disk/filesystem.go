package disk

import (
	"encoding/json"
	"os"
	"sort"
)

/*
A type returned by the create swarm option.
Contains information related to metadata such as filehash associated with filesize and other such things.
*/
type SwarmStorage struct {
	SwarmId    string            "swid"
	amountused uint64            "amtused"
	files      map[string]uint64 "files"
	fileordering []string "fileorder"
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
	i.files=make(map[string]uint64)
	err = os.Mkdir(swarmid, os.ModeDir|os.ModePerm)
	if err != nil {
		if os.IsExist(err) {
			meta, e := os.Open(swarmid + ".conf")
			if e != nil {
			}
			defer meta.Close()
			c := json.NewDecoder(meta)
			if e != nil {
				print(e.Error())
			}
			if err = c.Decode(&i); e != nil {

			}
		}
	}
	r = &i
	return
}

func (r SwarmStorage) CreateFile(filehash string, length uint64) (written int64, err error) {
	file, err := os.Create(r.SwarmId + string(os.PathSeparator) + filehash)

	if err != nil&&os.IsExist(err) {
		//in which case, it should be safe to ignore the error
		err = nil
		if r.files[filehash]==length{
			return
		}
	}
	defer file.Close()
	err = file.Truncate(int64(length))
	if _,ok:=r.files[filehash];!ok{
		r.fileordering=append(r.fileordering,filehash)
		sort.Strings(r.fileordering)
	}
	r.files[filehash] = uint64(length)
	written = int64(length)
	return
}
func (r SwarmStorage) FileExists(filehash string) bool{
	_,ok:=r.files[filehash]
	return ok
}

func (r SwarmStorage) DeleteFile(filehash string) error {
	size, err := os.Stat(r.getFileName(filehash))
	if err == nil {
		r.amountused -= uint64(size.Size())
		err = os.Remove(r.SwarmId + string(os.PathSeparator) + filehash)
	}
	r.files[filehash] = uint64(0)
	return err
}

func (r SwarmStorage) WriteFile(filehash string, start uint64, data []byte) error {
	path := r.SwarmId + string(os.PathSeparator) + filehash
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	size, ok := r.files[filehash]
	if uint64(start)+uint64(len(data)) >= size && ok {
		r.amountused += uint64(start) + uint64(uint64(len(data))-size)
		r.files[filehash] = uint64(start) + uint64(len(data))
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
func (r SwarmStorage) SaveSwarm() {
	s,err:=os.Create(r.SwarmId+".conf")
	if err!=nil&&os.IsExist(err){
		s, err = os.Open(r.SwarmId + ".conf")
	}
	defer s.Close()
	js := json.NewEncoder(s)
	if err = js.Encode(&r); err != nil {
		print("From SaveSwarm")
		print(err.Error())
	}

}
func (r SwarmStorage) GetRandomByte(index uint64) byte{
	var u uint64
	c:=uint64(0)
	v:=""
	for i :=range r.fileordering{
		var d=r.fileordering[i]
		if u+r.files[d]>=index{
			c=index-u+index
			v=d
			break
		}
		u+=r.files[d]
	}
	if v==""{
		return 0
	}
	b:=[]byte{0}
	r.ReadFile(v,c,b)
	return b[0]
}
