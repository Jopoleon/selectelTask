package egorServer

import (
	"crypto/tls"
	"errors"
	"io/ioutil"
	"os"

	"log"

	"net/http"

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
	cntlist    []ContainerInfo
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
	// This will remain the official name for now
	return "Welcome on Egor's FTP server for Swift API \n Enter your contract ID and then your password", nil
}

// AuthUser authenticates the user and selects an handling driver
func (driver *MainDriver) AuthUser(cc server.ClientContext, user, pass string) (server.ClientHandlingDriver, error) {

	authURL := "https://auth.selcdn.ru/"
	log.Println("...........////", authURL, user, pass)
	client := &http.Client{}
	r, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return nil, errors.New("http.NewRequest error: " + err.Error())
	}

	//$ curl -i https://auth.selcdn.ru/ \
	//-H "X-Auth-User:user" \
	//-H "X-Auth-Key:password"
	r.Header.Add("X-Auth-User", user)
	r.Header.Add("X-Auth-Key", pass)
	log.Println(authURL+" Request: ", r)
	resp, err := client.Do(r)
	if err != nil {
		return nil, errors.New("http.Request.Do error: " + err.Error())
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	log.Println(authURL, " Response: ", resp)
	if resp.Header.Get("X-Auth-Token") != "" && resp.Header.Get("X-Expire-Auth-Token") != "" && resp.Header.Get("X-Storage-Token") != "" && resp.Header.Get("X-Storage-Url") != "" {
		driver.authInfo.AuthToken = resp.Header.Get("X-Auth-Token")
		driver.authInfo.ExpireAuthToken = resp.Header.Get("X-Expire-Auth-Token")
		driver.authInfo.StorageUrl = resp.Header.Get("X-Storage-Url")
		driver.authInfo.StorageToken = resp.Header.Get("X-Storage-Token")
	}
	log.Printf(" %s \n\n Response headers: %+v", authURL, driver.authInfo)
	//client := &http.Client{}
	r, err = http.NewRequest("GET", driver.authInfo.StorageUrl, nil)
	if err != nil {
		return nil, errors.New("http.NewRequest error: " + err.Error())
	}

	//$ curl -i https://auth.selcdn.ru/ \
	//-H "X-Auth-User:user" \
	//-H "X-Auth-Key:password"
	//r.Header.Add("X-Auth-User", user)
	r.Header.Add("X-Auth-Token", driver.authInfo.AuthToken)
	log.Println(driver.authInfo.StorageUrl+" Request: ", r)
	resp, err = client.Do(r)
	if err != nil {
		return nil, errors.New("http.Request.Do error: " + err.Error())
	}
	if resp != nil {
		defer resp.Body.Close()
	}
	log.Println(driver.authInfo.StorageUrl, " Response: ", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("ioutil.ReadAll error: " + err.Error())
	}
	log.Println(driver.authInfo.StorageUrl, " response Body: ", string(body))
	//driver.cntlist =
	driver.baseDir = "/"
	//X-Auth-Token:[a529d07fc55820e3721bf6e5aafc16ac]
	//X-Expirs-Auth-Token:[95775]
	//X-Storage-Token:[a529d07fc55820e3721bf6e5aafc16ac]
	//Date:[Thu, 08 Jun 2017 18:32:34 GMT]
	//Access-Control-Allow-Origin:[*]
	//X-Storage-Url:[https://236505.selcdn.ru/]
	//Милосердов Егор
	//Номер договора (логин): 68462
	//jzDVODFQ
	return driver, nil
}

// ChangeDirectory changes the current working directory
func (driver *MainDriver) ChangeDirectory(cc server.ClientContext, directory string) error {
	//log.Println("ChangeDirectory ", directory)
	//if directory == "" {
	//	return errors.New("Directory cannot be empty")
	//}
	//directory = strings.Trim(directory, "/")
	//if contains(driver.containers, directory) {
	//	driver.baseDir = driver.baseDir + directory
	//	return nil
	//} else {
	//	return errors.New("No such container name on list")
	//}
	return nil
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// MakeDirectory creates a directory
func (driver *MainDriver) MakeDirectory(cc server.ClientContext, directory string) error {
	return os.Mkdir(driver.baseDir+directory, 0777)
}

// ListFiles lists the files of a directory
func (driver *MainDriver) ListFiles(cc server.ClientContext) ([]os.FileInfo, error) {
	//Милосердов Егор
	//Номер договора (логин): 68462
	//jzDVODFQ
	log.Println("ListFiles path: ", cc.Path())

	//if current ftp working directory is /, then get list of available containers
	if cc.Path() == "/" {
		client := &http.Client{}
		r, err := http.NewRequest("GET", driver.authInfo.StorageUrl, nil)
		if err != nil {
			return nil, errors.New("http.NewRequest error: " + err.Error())
		}
		r.Header.Add("X-Auth-Token", driver.authInfo.AuthToken)
		log.Println(driver.authInfo.StorageUrl+" Request: ", r)
		resp, err := client.Do(r)
		if err != nil {
			return nil, errors.New("http.Request.Do error: " + err.Error())
		}
		if resp != nil {
			defer resp.Body.Close()
		}
		log.Println(driver.authInfo.StorageUrl, " Response: ", resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.New("ioutil.ReadAll error: " + err.Error())
		}
		log.Println(driver.authInfo.StorageUrl, " response Body: ", string(body))
		var s = strings.Split(string(body), "\n")
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

	} else {
		client := &http.Client{}
		r, err := http.NewRequest("GET", driver.authInfo.StorageUrl+cc.Path(), nil)
		if err != nil {
			return nil, errors.New("http.NewRequest error: " + err.Error())
		}
		r.Header.Add("X-Auth-Token", driver.authInfo.AuthToken)
		log.Println(driver.authInfo.StorageUrl+" Request: ", r)
		resp, err := client.Do(r)
		if err != nil {
			return nil, errors.New("http.Request.Do error: " + err.Error())
		}
		if resp != nil {
			defer resp.Body.Close()
		}
		log.Println(driver.authInfo.StorageUrl, " Response: ", resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.New("ioutil.ReadAll error: " + err.Error())
		}
		if resp.Status == "404" || strings.Contains(string(body), "Not Found") {
			return nil, errors.New("The resource " + cc.Path() + " could not be found to list. ")
		}

		log.Println(driver.authInfo.StorageUrl+cc.Path(), " response Body: ", string(body))
		var s = strings.Split(string(body), "\n")
		var filelist = s[:len(s)-1]
		log.Println("List of files in "+cc.Path()+": ", filelist)
		fileinfo := make([]os.FileInfo, 0)
		for _, c := range filelist {
			fileinfo = append(fileinfo, virtualFileInfo{
				name: c,
				mode: os.FileMode(0666),
				size: 1024,
			})
		}
		return fileinfo, nil
	}
	return nil, nil
}

// UserLeft is called when the user disconnects, even if he never authenticated
func (driver *MainDriver) UserLeft(cc server.ClientContext) {

}

// OpenFile opens a file in 3 possible modes: read, write, appending write (use appropriate flags)
func (driver *MainDriver) OpenFile(cc server.ClientContext, path string, flag int) (server.FileStream, error) {
	log.Println("OpenFile used with path: ", path, " and flag: ", flag)
	log.Println("OpenFile used with cc.Path(): ", cc.Path())
	//Милосердов Егор
	//Номер договора (логин): 68462
	//jzDVODFQ

	//Загрузка файла
	//PUT запрос X-Storage-Url/container_name/path_to_file.
	//Flag == 1 is WriteOnly mode
	if flag == 1 {
		client := &http.Client{}
		log.Println("path Before String.Split: ", path)
		f1 := strings.Split(path, "/")
		//strings.Replace(path, "\\", "", -1)
		log.Println("path after String.Split: ", f1, " and it's Len():", len(f1))

		f2 := f1[len(f1)-1]
		log.Println("path after f1[len(f1)-1]: ", f2)
		file, err := os.Open(strings.Trim(f2, "/"))
		defer file.Close()
		if err != nil {
			return nil, errors.New(" os.Open(path) error: " + err.Error())
		}
		r, err := http.NewRequest("PUT", driver.authInfo.StorageUrl+path, nil)
		if err != nil {
			return nil, errors.New("http.NewRequest error: " + err.Error())
		}
		r.Header.Add("X-Auth-Token", driver.authInfo.AuthToken)
		log.Println(driver.authInfo.StorageUrl+" Request: ", r)
		resp, err := client.Do(r)
		if err != nil {
			return nil, errors.New("http.Request.Do error: " + err.Error())
		}
		if resp != nil {
			defer resp.Body.Close()
		}
		log.Println(driver.authInfo.StorageUrl+cc.Path()+path, " Response: ", resp)
		return nil, nil
	}
	//Чтение файла
	//GET запрос на адрес X-Storage-Url/container_name/path_to_file.
	//Flag == 0 is ReadOnly mode
	if flag == 0 {
		client := &http.Client{}
		//log.Println("path Before String.Split: ", path)
		//f1 := strings.Split(path, "/")
		////strings.Replace(path, "\\", "", -1)
		//log.Println("path after String.Split: ", f1, " and it's Len():", len(f1))
		//
		//f2 := f1[len(f1)-1]
		//log.Println("path after f1[len(f1)-1]: ", f2)
		path = strings.Trim(path, "/")

		r, err := http.NewRequest("GET", driver.authInfo.StorageUrl+path, nil)
		if err != nil {
			return nil, errors.New("http.NewRequest error: " + err.Error())
		}
		r.Header.Add("X-Auth-Token", driver.authInfo.AuthToken)
		log.Println(driver.authInfo.StorageUrl+path, " Request: ", r)
		resp, err := client.Do(r)
		if err != nil {
			return nil, errors.New("http.Request.Do error: " + err.Error())
		}
		if resp != nil {
			defer resp.Body.Close()
		}
		log.Println(driver.authInfo.StorageUrl+cc.Path()+path, " Response: ", resp)
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.New("ioutil.ReadAll(resp.Body) error: " + err.Error())
		}
		if resp.Status == "404" || strings.Contains(string(body), "Not Found") {
			return nil, errors.New("The file " + path + " not found. ")
		}

		f1 := strings.Split(path, "/")
		//strings.Replace(path, "\\", "", -1)
		log.Println("path after String.Split: ", f1, " and it's Len():", len(f1))

		f2 := f1[len(f1)-1]
		log.Println("last element of path path[len(path)-1]: ", f2)
		file, err := os.Create(f2)
		if err != nil {
			return nil, errors.New("Cannot create file " + err.Error())
		}
		//func (f *File) Write(b []byte) (n int, err error)
		_, err = file.Write(body)
		if err != nil {
			return nil, errors.New("Cannot write to file " + err.Error())
		}
		file.Close()
		//err = ioutil.WriteFile(f2, body, 0777)
		//if err != nil {
		//	return nil, errors.New("ioutil.WriteFile(path, body, 0644) error: " + err.Error())
		//}
		return os.OpenFile(f2, flag, 0666)
		//return nil, nil
	}
	return nil, nil
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
