package filecache

import (
	"fmt"
	"testing"
	"time"
)

func TestKeyToFilename(t *testing.T) {
	fmt.Println(keyToFilename("test_key", 101010))
	fmt.Println(keyToFilename("test_key", 1111101010))
}

func TestNewDefault(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	fmt.Println(c)
}

func TestFileCache_Has(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	has := c.Has("test_key")
	fmt.Println(has)
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)
	has = c.Has("test_key")
	fmt.Println(has)
}

func TestFileCache_Set(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)
	c.Set("test_key2", []string{
		"1", "2", "3",
	}, time.Hour)
	c.Set("test_key3", time.Now(), time.Hour)
}

func TestFileCache_Set2(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)
	c.Set("test_key", []string{
		"1", "2", "3",
	}, time.Hour)
	c.Set("test_key", time.Now(), time.Hour)
}

func TestFileCache_Delete(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)
	v, err := c.Delete("test_key")
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(v)
}

func TestFileCache_Add(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	added, err := c.Add("test_key", value, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(added)
	t.Log(c.Has("test_key"))
}

func TestFileCache_Get(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	added, err := c.Add("test_key", value, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("added: ", added)
	v, found := c.Get("test_key")
	t.Log("found: ", found)
	if found {
		t.Log(v)
	}
}

func TestFileCache_Replace(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)

	v, found := c.Get("test_key")
	t.Log("found: ", found)
	if found {
		t.Log(v)
	}

	replaced, err := c.Replace("test_key", map[string]string{
		"a1": "a1",
	}, time.Hour)

	t.Log("replaced: ", replaced)
	if err != nil {
		t.Log(err)
	}
}

func TestFileCache_Clear(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	value := map[string]string{
		"key1": "value1",
		"key2": "value1",
		"key3": "value1",
		"key4": "value1",
		"key5": "value1",
	}
	c.Set("test_key", value, time.Hour)
	c.Set("test_key1", value, time.Hour)
	c.Clear()
}

func TestFileCache_OnAddHook(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	c.OnAddHook(func(k string, v interface{}) {
		fmt.Println("defined add hook")
		fmt.Println("key: ", k, "value: ", v)
	})
	added, err := c.Add("test", 11111, time.Hour)
	t.Log(added)

	if err != nil {
		t.Fatal(err)
	}
}

func TestFileCache_ItemCount(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	added, err := c.Add("test", 11111, time.Hour)
	t.Log(added)

	if err != nil {
		t.Fatal(err)
	}
	t.Log(c.ItemCount())
}

func TestFileCache_DeleteExpired(t *testing.T) {
	c := New("D:\\workspace\\github.com\\a2htray\\filecache\\testdata")
	c.Set("k1", 1, -time.Hour)
	c.Set("k2", 2, time.Hour)
	deleted, err := c.DeleteExpired()
	t.Log(deleted)

	if err != nil {
		t.Fatal(err)
	}
}

func TestNewWithMonitor(t *testing.T) {
	fc := NewWithMonitor("D:\\workspace\\github.com\\a2htray\\filecache\\testdata", time.Second*10)
	fc.Set("key1", 11111, time.Second*5)
	for {
		t.Log(fc.items)
		continue
	}
}
