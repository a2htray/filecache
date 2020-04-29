package filecache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)

var (
	default2Hook = func(k string, v interface{}) {
		fmt.Printf("key: %s, value: %s\n", k, v)
	}
	default3Hook = func(k string, oldV interface{}, newV interface{}) {
		fmt.Printf("key: %s, old value: %s, new value: %s\n", k, oldV, newV)
	}
)

var hash = md5.New()

func keyToFilename(key string, expired int64) string {
	hash.Reset()
	hash.Write([]byte(key))
	return fmt.Sprintf("%s_%d", hex.EncodeToString(hash.Sum(nil)), expired)
}

type Cache interface {
	Set(k string, x interface{}, d time.Duration)
	Delete(k string) (interface{}, error)
	Add(k string, x interface{}, d time.Duration) (bool, error)
	Get(k string) (interface{}, bool)
	Replace(k string, x interface{}, d time.Duration) (bool, error)
	Has(k string) bool
	Clear() bool
	ItemCount() int
}

type Item struct {
	expiration int64
	filename   string
}

func (i Item) IsExpired() bool {
	return time.Now().UnixNano() > i.expiration
}

func (i Item) Filename() string {
	return i.filename
}

type FileCache struct {
	mutex         sync.RWMutex
	items         map[string]Item
	onSetHook     func(k string, v interface{})
	onDeleteHook  func(k string, v interface{})
	onAddHook     func(k string, v interface{})
	onReplaceHook func(k string, originV interface{}, newV interface{})
	dir           string
	monitor       *monitor
}

func (f *FileCache) OnSetHook(onSetHook func(k string, v interface{})) {
	f.onSetHook = onSetHook
}

func (f *FileCache) OnDeleteHook(onDeleteHook func(k string, v interface{})) {
	f.onDeleteHook = onDeleteHook
}

func (f *FileCache) OnAddHook(onAddHook func(k string, v interface{})) {
	f.onAddHook = onAddHook
}

func (f *FileCache) OnReplaceHook(onReplaceHook func(k string, originV interface{}, newV interface{})) {
	f.onReplaceHook = onReplaceHook
}

func (f *FileCache) Set(k string, x interface{}, d time.Duration) {
	has := f.Has(k)

	f.mutex.Lock()
	defer f.mutex.Unlock()

	if has {
		err := os.Remove(path.Join(f.dir, f.items[k].filename))
		if err != nil {
			panic(err)
		}

		delete(f.items, k)
	}

	f.set(k, x, d)
	f.onSetHook(k, x)
}

func (f *FileCache) set(k string, x interface{}, d time.Duration) {
	expiration := time.Now().Add(d).UnixNano()
	filename := keyToFilename(k, expiration)
	file, err := os.Create(path.Join(f.dir, filename))
	defer file.Close()

	if err != nil {
		panic(err)
	}

	content, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(string(content))
	if err != nil {
		panic(err)
	}

	f.items[k] = Item{
		filename:   filename,
		expiration: expiration,
	}
}

func (f *FileCache) Has(k string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if _, ok := f.items[k]; ok {
		return true
	}

	return false
}

func (f *FileCache) Delete(k string) (interface{}, error) {
	has := f.Has(k)
	if has == false {
		return nil, errors.New("item is not existed")
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	v, err := f.delete(k)
	if err != nil {
		return nil, err
	}

	f.onDeleteHook(k, v)

	return v, nil
}

func (f *FileCache) delete(k string) (interface{}, error) {
	filename := path.Join(f.dir, f.items[k].filename)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var v interface{}
	err = json.Unmarshal(content, &v)
	if err != nil {
		return nil, err
	}

	err = os.Remove(path.Join(f.dir, f.items[k].filename))
	if err != nil {
		return nil, err
	}

	delete(f.items, k)
	return v, nil
}

func (f *FileCache) Add(k string, x interface{}, d time.Duration) (bool, error) {
	has := f.Has(k)
	if has {
		return false, errors.New("item is existed")
	}
	expiration := time.Now().Add(d).UnixNano()
	filename := keyToFilename(k, expiration)
	content, err := json.Marshal(x)
	if err != nil {
		return false, err
	}

	f.mutex.Lock()
	defer f.mutex.Unlock()

	file, err := os.Create(path.Join(f.dir, filename))
	if err != nil {
		return false, err
	}

	_, err = file.WriteString(string(content))
	if err != nil {
		return false, err
	}

	f.items[k] = Item{
		expiration: expiration,
		filename:   filename,
	}

	f.onAddHook(k, x)

	return true, nil
}

func (f *FileCache) Get(k string) (interface{}, bool) {
	has := f.Has(k)
	if has == false {
		return nil, false
	}

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	item := f.items[k]
	if item.IsExpired() {
		_, _ = f.delete(k)
		return nil, false
	}

	filename := path.Join(f.dir, f.items[k].filename)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, false
	}

	var v interface{}
	err = json.Unmarshal(content, &v)
	if err != nil {
		return nil, false
	}

	return v, true
}

func (f *FileCache) Replace(k string, x interface{}, d time.Duration) (bool, error) {
	has := f.Has(k)
	if has == false {
		return false, errors.New("item is not existed")
	}

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	oldV, err := f.delete(k)
	if err != nil {
		return false, err
	}

	f.set(k, x, d)
	f.onReplaceHook(k, oldV, x)

	return true, nil
}

func (f *FileCache) Clear() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for k, _ := range f.items {
		_, err := f.delete(k)
		if err != nil {
			return false
		}
	}

	f.items = map[string]Item{}
	return true
}

func (f *FileCache) ItemCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	return len(f.items)
}

func (f *FileCache) DeleteExpired() (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for k, item := range f.items {
		if item.IsExpired() {
			_, err := f.delete(k)
			if err != nil {
				return false, err
			}
		}
	}

	return true, nil
}

func New(dir string) *FileCache {
	return &FileCache{
		onSetHook:     default2Hook,
		onDeleteHook:  default2Hook,
		onAddHook:     default2Hook,
		onReplaceHook: default3Hook,
		dir:           dir,
		items:         make(map[string]Item),
	}
}

func NewWithMonitor(dir string, ci time.Duration) *FileCache {
	fc := &FileCache{
		onSetHook:     default2Hook,
		onDeleteHook:  default2Hook,
		onAddHook:     default2Hook,
		onReplaceHook: default3Hook,
		dir:           dir,
		items:         make(map[string]Item),
		monitor: &monitor{
			interval: ci,
			stop:     make(chan bool),
		},
	}
	go fc.monitor.Run(fc)

	return fc
}

type monitor struct {
	interval time.Duration
	stop     chan bool
}

func (m *monitor) Run(fc *FileCache) {
	ticker := time.NewTicker(m.interval)
	for {
		select {
		case <-ticker.C:
			_, _ = fc.DeleteExpired()
		case <-m.stop:
			ticker.Stop()
		}
	}
}

func (m *monitor) Stop(fc *FileCache) {
	fc.monitor.stop <- true
}
