package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/iansinnott/browser-gopher/pkg/config"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup your data",
	Long:  `This command will backup your database and search index.`,
	Run: func(cmd *cobra.Command, args []string) {
		backupPath := filepath.Join(config.Config.BackupDir, "backup_"+time.Now().Format("20060102150405"))
		err := os.MkdirAll(backupPath, 0755)
		if err != nil {
			fmt.Println("could not create backup dir", err)
			os.Exit(1)
		}

		// copy the config dir to the backup dir
		err = copyDir(config.Config.AppDataPath, backupPath)
		if err != nil {
			fmt.Println("could not copy config dir to backup dir", err)
			os.Exit(1)
		}

		fmt.Println("Backed up to:", backupPath)
	},
}

func copyDir(src, dst string) error {
	// get properties of source dir
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// create destination dir
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
