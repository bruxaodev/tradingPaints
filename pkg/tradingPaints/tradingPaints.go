package tradingPaints

import (
	"bytes"
	"compress/bzip2"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/bruxaodev/tradingPaints/internal/dataBase"
	"github.com/bruxaodev/tradingPaints/internal/schemas"

	"github.com/bruxaodev/go-logger"
	"gorm.io/gorm"
)

var (
	initialize = false
	log        = logger.New("TRADINGPAINTS", logger.Green)
)

type TPXML struct {
	XMLName xml.Name `xml:"TPXML"`
	Cars    Cars     `xml:"Cars"`
}

type Cars struct {
	XMLName xml.Name `xml:"Cars"`
	Car     []Car    `xml:"Car"`
}

type Car struct {
	Carid     string `xml:"carid"`
	File      string `xml:"file"`
	UserId    string `xml:"userid"`
	Directory string `xml:"directory"`
	Filesize  string `xml:"filesize"`
	Filesize2 string `xml:"filesize2"`
	Type      string `xml:"type"`
	Teamid    string `xml:"teamid"`
}

type Player struct {
	UserId  string `json:"userId"`
	CarName string `json:"carName"`
}

func Update(users []Player, force bool) error {
	if !initialize {
		return fmt.Errorf("TradingPaints not initialized")
	}
	// http request
	data, err := fetchPaints(users)
	if err != nil {
		log.Errorf("Error on fetch paints: %v\n", err)
		return err
	}
	// parse data
	cars, err := parseData(data)
	if err != nil {
		log.Errorf("Error on parse data: %v\n", err)
		return err
	}

	for _, car := range cars {
		update, err := updatePaintInDB(car)
		log.Infof("Update: %v\n", update)
		if err != nil {
			log.Errorf("Error on update paint in db: %v\n", err)
			downloadPaint(car)
			continue
		}
		if update || force {
			log.Infof("Downloading paint: %v\n", car.File)
			downloadPaint(car)
		}
	}
	return nil
}

func Init() error {

	log.Info("Starting TradingPaints")
	_, err := dataBase.NewSql()
	if err != nil {
		log.Error("Error on start TradingPaints", err)
		return err
	}
	initialize = true
	return nil
}

func fetchPaints(users []Player) ([]byte, error) {

	strData := "1=safety pcporsche911cup=0=0,"

	for index, user := range users {
		strData = fmt.Sprintf("%v%v=%v=%v=%v,", strData, user.UserId, user.CarName, 0, index+1)
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	defer w.Close()
	if err := w.WriteField("list", strData); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "http://dl.tradingpaints.com/fetch.php", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	//execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	//read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func downloadPaint(car Car) error {
	if car.File == "" {
		return nil
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}
	dir := filepath.Join(usr.HomeDir, "Documents", "iracing", "paint")
	tempDir := filepath.Join(usr.HomeDir, "AppData", "Local", "Temp")
	fileType := strings.Split(car.File, ".")
	log.Infof("Iniciando download: %v\n", car.File)

	resp, err := http.Get(car.File)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	helmet := car.Type == "helmet"
	suit := car.Type == "suit"
	isBz2 := fileType[len(fileType)-1] == "bz2"

	if _, err := os.Stat(filepath.Join(dir, car.Directory)); os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Join(dir, car.Directory), os.ModePerm)
		if err != nil {
			return err
		}
	}

	if isBz2 {
		tempPath := filepath.Join(tempDir, fmt.Sprintf("temp_%s_%s.%s", car.Directory, car.UserId, fileType[len(fileType)-1]))

		var outputFilePath string
		if helmet {
			outputFilePath = filepath.Join(dir, fmt.Sprintf("helmet_%s.%s", car.UserId, fileType[len(fileType)-2]))
		} else if suit {
			outputFilePath = filepath.Join(dir, fmt.Sprintf("suit_%s.%s", car.UserId, fileType[len(fileType)-2]))
		} else {
			outputFilePath = filepath.Join(dir, car.Directory, fmt.Sprintf("car_%s.%s", car.UserId, fileType[len(fileType)-2]))
		}

		err = os.WriteFile(tempPath, buffer, 0644)
		if err != nil {
			return err
		}

		tempFile, err := os.Open(tempPath)
		if err != nil {
			return err
		}
		defer tempFile.Close()

		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			return err
		}
		defer outputFile.Close()

		bz2Reader := bzip2.NewReader(tempFile)
		_, err = io.Copy(outputFile, bz2Reader)
		if err != nil {
			return err
		}

		log.Infof("Unzipping Completed : %v\n", outputFilePath)
		defer func() {
			if err := os.Remove(tempPath); err != nil {
				log.Errorf("Fail to delete temp file: %v\n", err)
			} else {
				log.Info("Success, temp file deleted.")
			}
		}()

	} else {
		var outputFilePath string
		if helmet {
			outputFilePath = filepath.Join(dir, fmt.Sprintf("helmet_%s.%s", car.UserId, fileType[len(fileType)-1]))
		} else if suit {
			outputFilePath = filepath.Join(dir, fmt.Sprintf("suit_%s.%s", car.UserId, fileType[len(fileType)-1]))
		} else {
			outputFilePath = filepath.Join(dir, car.Directory, fmt.Sprintf("car_%s.%s", car.UserId, fileType[len(fileType)-1]))
		}
		err = os.WriteFile(outputFilePath, buffer, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePaintInDB(paint Car) (bool, error) {
	_path := strings.Split(paint.File, "/")
	_file := strings.Split(_path[len(_path)-1], ".")
	fileId := _file[0]
	fileType := _file[len(_file)-1]
	db := dataBase.GetDb()

	mip := fileType == "mip"
	bz2 := fileType == "bz2"

	p, err := db.GetPaint(paint.UserId, paint.Directory)
	if err != nil {
		// if not found, add paint
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err := db.AddPaint(schemas.Paint{
				UserId:  paint.UserId,
				CarName: paint.Directory,
				FileId:  fileId,
				Bz2:     bz2,
				Mip:     mip,
			})
			if err != nil {
				log.Errorf("Error on add paint: %v\n", err)
				return false, err
			}
			return true, nil
		}
		log.Errorf("Error on get paint: %v\n", err)
		return false, err
	}

	if p.FileId != fileId {
		log.Infof("Updating paint: %v\n", fileId)
		if err := db.UpdateFileId(p, fileId); err != nil {
			log.Infof("Error on update paint: %v\n", err)
			return false, err
		}
		if err := db.UpdateMip(p, false); err != nil {
			log.Infof("Error on update paint: %v\n", err)
			return false, err
		}
		if err := db.UpdateBz2(p, false); err != nil {
			log.Infof("Error on update paint: %v\n", err)
			return false, err
		}
		p.Bz2 = false
		p.Mip = false
	}

	if !p.Bz2 && bz2 {
		log.Infof("Updating paint: %v\n", fileId)
		if err := db.UpdateBz2(p, true); err != nil {
			log.Infof("Error on update paint: %v\n", err)
			return false, err
		}
		return true, nil
	} else if !p.Mip && mip {
		log.Infof("Updating paint: %v\n", fileId)
		if err := db.UpdateMip(p, true); err != nil {
			log.Infof("Error on update paint: %v\n", err)
			return false, err
		}
		return true, nil
	}

	return false, nil
}
