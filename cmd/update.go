/*
Copyright © 2024 NAME HERE stanyhelberth@gmail.com
*/
package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/oracle/oci-go-sdk/example/helpers"
	"github.com/oracle/oci-go-sdk/v49/objectstorage"
	"github.com/spf13/cobra"
	"github.com/stanyzra/env-manager-v2/internal/utils"
	"gopkg.in/ini.v1"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update [flags] -p <project-name> -e <project-environment> (-n <name> -v <value>|--file <file>)",
	Short: "Update a environment variable or secret",
	Example: `env-manager-v2 update -p collection-back-end-v2.1 -e dev -t envs -n foo -v bar
env-manager-v2 update -p gollection-elastic -e homolog -t secrets -n moo -v baz
env-manager-v2 update -p collection-back-end-v2.1 -e dev -t envs -f /path/to/file`,
	Long: `Update a environment variable or secret from the environment file in OCI Object Storage.
The project and environment flags are required. You can update multiple environment variables or secrets
using a file. If the file flag is used, the name and valueflags is ignored. The file should be
in INI format WITH keys and values. If a environment variable or secret doesn't exists,
it will not be created. Use the create command to create a new environment variable or secret.`,
	Args: func(cmd *cobra.Command, args []string) error {
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("Error reading option flag: %v", err)
		}

		envName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Error reading option flag: %v", err)
		}

		envValue, err := cmd.Flags().GetString("value")
		if err != nil {
			log.Fatalf("Error reading option flag: %v", err)
		}

		if filePath == "" && (envName == "" || envValue == "") {
			return fmt.Errorf("requires --name and --value flags unless --file is used")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		project, err := utils.GetFlagString(cmd, "project", validProjects)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		envType, err := utils.GetFlagString(cmd, "type", validTypes)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		projEnvironment, err := utils.GetFlagString(cmd, "environment", validEnvs)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			log.Fatalf("Error reading option flag: %v", err)
		}

		fileName := fmt.Sprintf("%s_%s", projEnvironment, envType)

		configProvider, configFileName, err := utils.GetConfigProviderOCI()

		if err != nil {
			fmt.Println("Error getting config provider: ", err)
			return
		}

		ini_config, err := ini.Load(configFileName)
		if err != nil {
			fmt.Println("Error loading config file: ", err)
			return
		}

		sec := ini_config.Section("OCI")
		namespace := sec.Key("namespace").String()

		if err != nil {
			fmt.Println("Error getting config provider: ", err)
			return
		}

		client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(configProvider)
		helpers.FatalIfError(err)

		if filePath != "" {
			UpdateEnvFromFile(client, namespace, project, projEnvironment, envType, fileName, filePath)
		} else {
			envName, err := cmd.Flags().GetString("name")
			if err != nil {
				log.Fatalf("Error reading option flag: %v", err)
			}

			envValue, err := cmd.Flags().GetString("value")
			if err != nil {
				log.Fatalf("Error reading option flag: %v", err)
			}

			UpdateSingleEnv(client, namespace, project, projEnvironment, envType, envName, envValue, fileName)
		}
	},
}

func UpdateEnvironmentVariables(envFile *ini.File, userEnvsFile *ini.File) bool {
	isSaved := false
	for _, key := range userEnvsFile.Section("").Keys() {
		if !envFile.Section("").HasKey(key.Name()) {
			fmt.Printf("[WARNING] Key \"%s\" not found in environment file\n", key.Name())
			continue
		}

		isSaved = true
		envFile.Section("").Key(key.Name()).SetValue(key.Value())
	}

	return isSaved
}

func UpdateEnvFromFile(client objectstorage.ObjectStorageClient, namespace string, project string, projEnvironment string, envType string, fileName string, filePath string) {
	userEnvFile, err := ini.Load(filePath)
	if err != nil {
		fmt.Println("Error loading file: ", err)
		return
	}

	envFile, err := utils.GetEnvsFileAsIni(project, fileName, client, namespace, bucketName)

	if err != nil {
		fmt.Println("Error loading file: ", err)
		return
	}

	if UpdateEnvironmentVariables(envFile, userEnvFile) {
		utils.SaveEnvFile(client, namespace, project, fileName, envFile, bucketName)
		fmt.Printf("Environment variables saved in project \"%s\" in \"%s\" environment\n", project, projEnvironment)
	}
}

func UpdateSingleEnv(client objectstorage.ObjectStorageClient, namespace string, project string, projEnvironment string, envType string, envName string, envValue string, fileName string) {
	envFile, err := utils.GetEnvsFileAsIni(project, fileName, client, namespace, bucketName)

	if err != nil {
		fmt.Println("Error loading file: ", err)
		return
	}

	if !envFile.Section("").HasKey(envName) {
		fmt.Printf("[WARNING] Environment variable \"%s\" doesn't exists in project \"%s\" in \"%s\" environment\n", envName, project, projEnvironment)
		return
	}

	envFile.Section("").Key(envName).SetValue(envValue)
	utils.SaveEnvFile(client, namespace, project, fileName, envFile, bucketName)
	fmt.Printf("Environment variables saved in project \"%s\" in \"%s\" environment\n", project, projEnvironment)
}

func init() {
	rootCmd.AddCommand(updateCmd)

	validTypesStr := strings.Join(validTypes, ", ")
	validProjectsStr := strings.Join(validProjects, ", ")
	validProjectEnvsStr := strings.Join(validEnvs, ", ")

	updateCmd.Flags().StringP("type", "t", "envs", fmt.Sprintf("Specify the environment variable type (options: %s)", validTypesStr))
	updateCmd.Flags().StringP("project", "p", "", fmt.Sprintf("Specify the project name (options: %s)", validProjectsStr))
	updateCmd.Flags().StringP("environment", "e", "", fmt.Sprintf("Specify the project environment (options: %s)", validProjectEnvsStr))
	updateCmd.Flags().StringP("name", "n", "", "Specify the environment variable or secret name")
	updateCmd.Flags().StringP("value", "v", "", "Specify the environment variable or secret value")
	updateCmd.Flags().StringP("file", "f", "", "Specify a file containing a list of environment variables or secrets. The file should be in INI format.")

	updateCmd.MarkFlagRequired("project")
	updateCmd.MarkFlagRequired("environment")
}
