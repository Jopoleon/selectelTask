package app

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/fclairamb/ftpserver/server"
	"github.com/naoina/toml"
	"gopkg.in/inconshreveable/log15.v2"
)

// MainDriver defines a very basic serverftp driver
type MainDriver struct {
	baseDir    string
	tlsConfig  *tls.Config
	authInfo   AuthInfo
	containers []string
}

// AuthInfo struct uses for saving authorization information about user
type AuthInfo struct {
	ExpireAuthToken string
	StorageUrl      string
	AuthToken       string
	StorageToken    string
}

// ContainerInfo uses for unmarshaling info about containers
type ContainerInfo struct {
	Bytes   int    `json:"bytes"`
	Count   int    `json:"count"`
	Name    string `json:"name"`
	RxBytes int    `json:"rx_bytes"`
	TxBytes int    `json:"tx_bytes"`
	Type    string `json:"type"`
}

// WelcomeUser is called to send the very first welcome message
func (driver *MainDriver) WelcomeUser(cc server.ClientContext) (string, error) {
	cc.SetDebug(true)
	return "Welcome on Egor's FTP server for Swift API \n Enter your contract ID and then your password", nil
}

// AuthUser authenticates the user and selects an handling driver
func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {
	authURL := "https://auth.selcdn.ru/"
	//Милосердов Егор
	//Номер договора (логин): 68462
	//password : jzDVODFQ
	resp, err := MakeAuthRequest(user, pass, authURL)
	if err != nil {
		return nil, err
	}
	log15.Debug("MakeAuthRequest resp:", resp)
	driver.authInfo.AuthToken = resp.Header.Get("X-Auth-Token")
	driver.authInfo.ExpireAuthToken = resp.Header.Get("X-Expire-Auth-Token")
	driver.authInfo.StorageUrl = resp.Header.Get("X-Storage-Url")
	driver.authInfo.StorageToken = resp.Header.Get("X-Storage-Token")
	driver.baseDir = "/"
	return driver, nil
}

// ChangeDirectory changes the current working directory
func (driver *MainDriver) ChangeDirectory(cc server.ClientContext, directory string) error {

	return nil
}

// MakeDirectory creates a directory
func (driver *MainDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0777)
}

// ListFiles lists the files of a current ftp directory
func (driver *MainDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {

	log.Println("ListFiles path: ", cc.Path())
	resp, err := MakeStorageRequest("GET", driver.authInfo.StorageUrl+cc.Path(), driver.authInfo.AuthToken, nil)
	if err != nil {
		return nil, err
	}
	log.Println(driver.authInfo.StorageUrl, " response Body: ", string(resp))
	var s = strings.Split(string(resp), "\n")
	driver.containers = s[:len(s)-1]
	log.Println(" Containers list after parse: ", driver.containers)
	fileinfo := make([]os.FileInfo, 0)
	for _, c := range driver.containers {
		fileinfo = append(fileinfo, virtualFileInfo{
			name: c,
			mode: os.FileMode(0666),
			size: 1024,
		})
	}
	return fileinfo, nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {

}

// OpenFile loads file if PUT or SEND ftp command was used, or downloads file if ftp GET was used
func (driver *MainDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	log.Println("OpenFile used with path: ", path, " and flag: ", flag)
	log.Println("OpenFile used with cc.Path(): ", cc.Path())

	switch flag {
	//Загрузка файла
	//PUT запрос X-Storage-Url/container_name/path_to_file.
	//Flag == 1 is WriteOnly mode
	case 1:
		f1 := strings.Split(path, "/")
		//strings.Replace(path, "\\", "", -1)
		log15.Debug("path after String.Split: ", f1, " and it's Len():", len(f1))
		f2 := f1[len(f1)-1]
		log15.Debug("path after f1[len(f1)-1]: ", f2)
		file, err := os.Open(strings.Trim(f2, "/"))
		defer file.Close()
		if err != nil {
			return nil, errors.New(" os.Open(path) error: " + err.Error())
		}
		resp, err := MakeStorageRequest("PUT", driver.authInfo.StorageUrl+path, driver.authInfo.AuthToken, bufio.NewReader(file))
		if err != nil {
			return nil, err
		}
		log15.Debug("PUT response: ", string(resp))
		return os.OpenFile(f2, flag, 0666)
	//Чтение файла
	//GET запрос на адрес X-Storage-Url/container_name/path_to_file.
	//Flag == 0 is ReadOnly mode
	case 0:
		resp, err := MakeStorageRequest("GET", driver.authInfo.StorageUrl+path, driver.authInfo.AuthToken, nil)
		f1 := strings.Split(path, "/")
		log15.Debug("path after String.Split: ", f1, " and it's Len():", len(f1))
		f2 := f1[len(f1)-1]
		log.Println("last element of path path[len(path)-1]: ", f2)
		file, err := os.Create(f2)
		if err != nil {
			return nil, errors.New("Cannot create file " + err.Error())
		}
		_, err = file.Write(resp)
		if err != nil {
			return nil, errors.New("Cannot write to file " + err.Error())
		}
		file.Close()
		return os.OpenFile(f2, flag, 0666)
	default:
		return nil, errors.New("OpenFile error: Unknown FileMode flag")
	}
}

// GetFileInfo gets some info around a file or a directory
func (driver *MainDriver) GetFileInfo(cc server.ClientContext, path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// GetTLSConfig returns a TLS Certificate to use
func (driver *MainDriver) GetTLSConfig() (*tls.Config, error) {
	if driver.tlsConfig == nil {
		log15.Info("Loading certificate")
		if cert, err := tls.LoadX509KeyPair("sample/certs/mycert.crt", "sample/certs/mycert.key"); err == nil {
			driver.tlsConfig = &tls.Config{
				NextProtos:   []string{"ftp"},
				Certificates: []tls.Certificate{cert},
			}
		} else {
			return nil, err
		}
	}
	return driver.tlsConfig, nil
}

// CanAllocate gives the approval to allocate some data
func (driver *MainDriver) CanAllocate(cc server.ClientContext, size int) (bool, error) {
	return true, nil
}

// ChmodFile changes the attributes of the file
func (driver *MainDriver) ChmodFile(cc server.ClientContext, path string, mode os.FileMode) error {
	return nil
}

// DeleteFile deletes a file or a directory
func (driver *MainDriver) DeleteFile(cc server.ClientContext, path string) error {
	return nil
}

// RenameFile renames a file or a directory
func (driver *MainDriver) RenameFile(cc server.ClientContext, from, to string) error {
	return nil
}

// GetSettings returns some general settings around the server setup
func (driver *MainDriver) GetSettings() *server.Settings {
	f, err := os.Open("settings.toml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var config server.Settings
	if err := toml.Unmarshal(buf, &config); err != nil {
		panic(err)
	}

	// This is the new IP loading change coming from Ray
	if config.PublicHost == "" {
		log15.Debug("Fetching our external IP address...")
		if config.PublicHost, err = externalIP(); err != nil {
			log15.Warn("Couldn't fetch an external IP", "err", err)
		} else {
			log15.Debug("Fetched our external IP address", "ipAddress", config.PublicHost)
		}
	}

	return &config
}

// NewSampleDriver creates a sample driver
// Note: This is not a mistake. Interface can be pointers. There seems to be a lot of confusion around this in the
//       server_ftp original code.
func NewSampleDriver() *MainDriver {
	dir, err := ioutil.TempDir("", "ftpserver")
	if err != nil {
		log15.Error("Could not find a temporary dir", "err", err)
	}

	driver := &MainDriver{
		baseDir: dir,
	}
	os.MkdirAll(driver.baseDir, 0777)
	return driver
}
