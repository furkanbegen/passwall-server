package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/pass-wall/passwall-server/internal/encryption"
	"github.com/pass-wall/passwall-server/internal/storage"
	"github.com/pass-wall/passwall-server/model"
	"github.com/spf13/viper"
)

// BackupData ...
func BackupData(s storage.Store) error {
	backupFolder := viper.GetString("backup.folder")
	backupPath := fmt.Sprintf("%s/passwall-%s.bak", backupFolder, time.Now().Format("2006-01-02T15-04-05"))

	var logins []model.Login
	s.Find(&logins)
	logins = DecryptLoginPasswords(logins)

	// Struct to []byte vs. vs.
	loginBytes := new(bytes.Buffer)
	json.NewEncoder(loginBytes).Encode(logins)

	if _, err := os.Stat(backupFolder); os.IsNotExist(err) {
		//http://permissions-calculator.org/
		//0755 Commonly used on web servers. The owner can read, write, execute.
		//Everyone else can read and execute but not modify the file.
		os.Mkdir(backupFolder, 0755)
	} else if err == nil {
		// is exist folder
	} else {
		err := errors.New("Error occured while backuping data")
		return err
	}

	encryption.EncryptFile(backupPath, loginBytes.Bytes(), viper.GetString("server.passphrase"))

	rotateBackup()

	return nil
}

// Rotate backup files
func rotateBackup() error {
	backupRotation := viper.GetInt("backup.rotation")
	backupFolder := viper.GetString("backup.folder")

	backupFiles, err := GetBackupFiles()
	if err != nil {
		return err
	}

	if len(backupFiles) > backupRotation {
		sort.SliceStable(backupFiles, func(i, j int) bool {
			return backupFiles[i].ModTime().After(backupFiles[j].ModTime())
		})

		for _, file := range backupFiles[backupRotation:] {
			_ = os.Remove(filepath.Join(backupFolder, file.Name()))
		}
	}

	return nil
}

func GetBackupFiles() ([]os.FileInfo, error) {
	backupFolder := viper.GetString("backup.folder")

	files, err := ioutil.ReadDir(backupFolder)
	if err != nil {
		return nil, err
	}

	var backupFiles []os.FileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "passwall") && strings.HasSuffix(file.Name(), ".bak") {
			backupFiles = append(backupFiles, file)
		}
	}

	return backupFiles, nil
}
