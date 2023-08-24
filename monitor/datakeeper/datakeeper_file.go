package datakeeper

import (
	"bytes"
	"context"
	"fmt"
	"monitor/protocol"
	"monitor/storage"
	"monitor/utils"
	"os"
	"path/filepath"
	"time"
)

var _ utils.Keeper = &FileDataKeeper{}

type FileDataKeeper struct {
	storeFilePath string
}

func NewFileDataKeeper(ctx context.Context, storeFilePath string) *FileDataKeeper {
	return &FileDataKeeper{
		storeFilePath: storeFilePath,
	}
}

func (f *FileDataKeeper) Init(ctx context.Context) error {
	err := f.mkDir(ctx)
	if err != nil {
		return fmt.Errorf("make path dir fail %s", err)
	}
	err = f.readData(ctx)
	if err != nil {
		return fmt.Errorf("read file data fail %s", err)
	}
	go func() {
		f.writtingData(ctx)
	}()
	return nil
}

func (f *FileDataKeeper) ShutDown(ctx context.Context) {
	err := f.writeData(ctx)
	if err != nil {
		utils.Errorf("wirte data fail %s", err)
	}
}

func (f *FileDataKeeper) writtingData(ctx context.Context) {
	for {
		<-time.After(time.Minute)
		err := f.writeData(ctx)
		if err != nil {
			utils.Warnf("wirte data fail %s", err)
		}
	}
}

func (f *FileDataKeeper) writeData(ctx context.Context) error {
	for key, store := range storage.AllDatasStorage {
		datas := store.LoadAll()
		dataBody := []byte{}
		for _, data := range datas {
			dataBody = append(dataBody, data.(protocol.DataConvert).ToFileData()...)
			dataBody = append(dataBody, []byte("\n")...)
		}
		err := f.updateFile(ctx, key, dataBody)
		if err != nil {
			return fmt.Errorf("update file %s fail %s", key, err)
		}
	}
	return nil
}

func (f *FileDataKeeper) readData(ctx context.Context) error {
	for key, store := range storage.AllDatasStorage {
		body, err := f.fetchFile(ctx, key)
		if err != nil {
			return fmt.Errorf("fetch file %s fail %s", key, err)
		}
		lines := bytes.Split(body, []byte("\n"))
		var (
			ks = []interface{}{}
			vs = []interface{}{}
		)
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			k, v, err := protocol.FileDataToStorage(key, line)
			if err != nil {
				return fmt.Errorf("file data to struct fail %s %s", string(line), err)
			}
			ks = append(ks, k)
			vs = append(vs, v)
		}
		store.Store(ks, vs)
	}
	return nil
}

func (f *FileDataKeeper) appendFile(ctx context.Context, fileName string, body []byte) error {
	filePath := filepath.Join(f.storeFilePath, fileName)
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file fail %s", err)
	}
	defer file.Close()
	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("write file fail %s", err)
	}
	return nil
}

func (f *FileDataKeeper) updateFile(ctx context.Context, fileName string, body []byte) error {
	filePath := filepath.Join(f.storeFilePath, fileName)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("open file fail %s", err)
	}
	defer file.Close()
	_, err = file.Write(body)
	if err != nil {
		return fmt.Errorf("write file fail %s", err)
	}
	return nil
}

func (f *FileDataKeeper) fetchFile(ctx context.Context, fileName string) ([]byte, error) {
	filePath := filepath.Join(f.storeFilePath, fileName)
	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("read file fail %s", err)
	}
	return b, nil
}

func (f *FileDataKeeper) mkDir(ctx context.Context) error {
	if f.storeFilePath == "" {
		return fmt.Errorf("storage path empty")
	}
	return os.MkdirAll(f.storeFilePath, os.ModePerm)
}
