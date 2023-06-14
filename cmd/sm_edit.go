package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"secret-edit/internal/services"
	"strings"
	"syscall"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/yaml"
)

func init() {

}

func smEdit(cmd *cobra.Command, args []string) {
	projectID := cmd.Flag("project").Value.String()
	if projectID == "" {
		logrus.Errorf("project flag is required")
		return
	}

	if cmd.Flag("list").Value.String() == "true" {
		err := listSecrets(projectID)
		if err != nil {
			logrus.Errorf("failed to list secrets: %v", err)
		}
		return
	}

	if len(args) < 1 {
		logrus.Errorf("Secret ID is required")
		return
	}
	secretID := args[0]

	svc, err := services.ReturnGCPSMService(projectID)
	if err != nil {
		logrus.Errorf("failed to create SM service: %v", err)
		return
	}

	secretInfo, err := svc.GetSecretInfo(secretID)
	if err != nil {
		logrus.Errorf("failed to get secret info: %v", err)
		return
	}

	secretData, err := svc.GetSecretData(secretID, "latest")
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				if secretInfo != nil {
					logrus.Warnf("no versions found, creating new version for secret %q", secretID)
					time.Sleep(3 * time.Second)
					secretData = []byte{}
				} else {
					logrus.Errorf("secret %q not found", secretID)
				}
			default:
				logrus.Errorf("secret %q not found", secretID)
				return
			}
		}
	}

	tempFile, err := os.CreateTemp("", "scts-*.yml")
	if err != nil {
		logrus.Errorf("failed to create temp file: %v", err)
		return
	}
	defer cleanup(tempFile.Name())

	terminationCh := make(chan os.Signal, 1)
	signal.Notify(terminationCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-terminationCh
		cleanup(tempFile.Name())
		os.Exit(1)
	}()

	yamlSecretData, err := yaml.JSONToYAML(secretData)
	if err != nil {
		logrus.Errorf("failed to convert JSON to YAML: %v", err)
		return
	}

	_, err = tempFile.Write(yamlSecretData)
	if err != nil {
		logrus.Errorf("failed to write YAML to temp file: %v", err)
		return
	}
	if err := tempFile.Close(); err != nil {
		logrus.Errorf("failed to close temp file: %v", err)
		return
	}

	newYamlSecretData := []byte{}
	for {
		if err := runEditor(tempFile.Name()); err != nil {
			logrus.Errorf("failed to run editor: %v", err)
			return
		}

		tempFileData, err := os.ReadFile(tempFile.Name())
		if err != nil {
			logrus.Errorf("failed to read temp file: %v", err)
			return
		}
		newYamlSecretData, err = yaml.JSONToYAML(tempFileData)
		if err != nil {
			logrus.Errorf("failed to convert edited YAML to JSON: %v", err)
			logrus.Info("Do you want to fix the error? (y/N): ")
			var input string
			fmt.Scanln(&input)
			if input == "y" || input == "Y" {
				continue
			} else if input == "n" || input == "N" {
				logrus.Warn("Operation aborted.")
				return
			} else {
				logrus.Warn("Invalid input. Please enter 'y' or 'n'.")
			}
		}
		break
	}

	if bytes.Equal(yamlSecretData, newYamlSecretData) {
		logrus.Info("No changes were made, ignoring...")
	} else {
		logrus.Infof("Changes detected, creating new secret version for %s", secretID)
		newJSONSecretData, err := yaml.YAMLToJSON(newYamlSecretData)
		if err != nil {
			logrus.Errorf("failed to convert edited YAML to JSON: %v", err)
			return
		}

		err = svc.CreateSecretVersion(secretID, bytes.NewReader(newJSONSecretData))
		if err != nil {
			logrus.Errorf("failed to add new secret version: %v", err)
			return
		}

	}
	logrus.Info("Done")
}

func listSecrets(projectID string) error {

	svc, err := services.ReturnGCPSMService(projectID)
	if err != nil {
		return err
	}

	secretList, err := svc.ListSecrets()
	if err != nil {
		return err
	}

	tablewriter := table.NewWriter()
	tablewriter.AppendHeader(table.Row{"Name", "Labels", "Created"})
	tablewriter.SetStyle(table.StyleLight)
	for _, secret := range secretList {
		name := strings.SplitN(secret.GetName(), "/", 4)[3]
		labels := secret.GetLabels()
		created := secret.GetCreateTime().AsTime().Format("2006-01-02T15:04:05")

		// transform map labels to json
		labelsJson, _ := json.Marshal(labels)

		row := table.Row{name, string(labelsJson), created}
		tablewriter.AppendRow(row)
	}

	fmt.Println(tablewriter.Render())
	return nil

}

func cleanup(filename string) {
	logrus.Debug("Cleaning up", filename)
	err := os.Remove(filename)
	if err != nil {
		logrus.Error("Error cleaning up temporary file:", err)
	}
}

func runEditor(filename string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	return runCommand(editor, filename)
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
