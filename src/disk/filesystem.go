package disk

import (
	"os"
	"bytes"
)

type SwarmStorage struct{
	SwarmId string
	amountused uint64
}

func  CreateSwarmSystem(swarmid string) *SwarmStorage,*error{
	var i SwarmId
	i.SwarmId=swarmid
	i.amountused=0
	err:=os.mkdir(swarmid,os.ModeDir)
	if err{
		return nil,err
	}
	return *i,nil
}

func (r SwarmStorage) CreateFile(filehash string, length uint64) error{
	file,err:=os.Create(r.SwarmId)

	if err==nil{
	defer file.Close()
		b:=[]byte{0}
		for i :=range length{
			_,_=file.Write(b)
		}
	}
	return err
}

func (r SwarmStorage) WriteFile(filehash string, start uint64, data []byte) error{
	path:=r.swarmid+os.PathSeparator+filehash
	file,err=os.Open(path)
	if err!=nil{
		panic("File could not be opened")
	}
	file.WriteAt(byte,start)
	file.Close()
}


















