package main

import (
	"path/filepath"
	"net/http"
	"encoding/json"
	"io/ioutil"

    "strings"
    "strconv"
	"log"
	"os"
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/awsutil"

	yaml "github.com/goccy/go-yaml"
)

func LoadSetting() *Setting {
    pwd, err := os.Getwd()
    if err != nil {
        log.Println(err)
    }

    if _, err := os.Stat(pwd + "/coral"); os.IsNotExist(err) {
		os.MkdirAll(pwd + "/coral", os.ModePerm)
    }
    
    if _, err := os.Stat(pwd + "/.build/sessions/image"); os.IsNotExist(err) {
		os.MkdirAll(pwd + "/.build/sessions", os.ModePerm)
    }

    return &Setting{
        UserAgent: getEnvString("USER_AGENT", "nemo/1.0.0"),
        RandMin: getEnvInt("RAND_MIN", 2),
        RandMax: getEnvInt("RAND_MAX", 4),
        LimitRandMax: getEnvInt("LIMIT_RAND_MAX", 5),
        SessionsDir: getEnvString("SESSION_DIR", pwd),
        CoralDir: getEnvString("CORAL_DIR", pwd + "/coral"),
        BuildDir: getEnvString("BUILD_DIR", pwd + "/.build"),
        AwsS3RegionName: getEnvString("AWS_S3_REGION_NAME", ""),
        AwsS3Dir: getEnvString("AWS_S3_DIR", ""),
        AwsS3EndpointUrl: getEnvString("AWS_S3_ENDPOINT_URL", ""),
        AwsStorageBucketName: getEnvString("AWS_STORAGE_BUCKET_NAME", ""),
        AwsSecretAccessKey: getEnvString("AWS_SECRET_ACCESS_KEY", ""),
        AwsAccessKeyId: getEnvString("AWS_ACCESS_KEY_ID", ""),
    }
}


func between(value string, a string, b string) string {
    // Get substring between two strings.
    posFirst := strings.Index(value, a)
    if posFirst == -1 {
        return ""
    }
    posLast := strings.Index(value, b)
    if posLast == -1 {
        return ""
    }
    posFirstAdjusted := posFirst + len(a)
    if posFirstAdjusted >= posLast {
        return ""
    }
    return value[posFirstAdjusted:posLast]
}

func after(value string, a string) string {
    // Get substring after a string.
    pos := strings.LastIndex(value, a)
    if pos == -1 {
        return ""
    }
    adjustedPos := pos + len(a)
    if adjustedPos >= len(value) {
        return ""
    }
    return value[adjustedPos:len(value)]
}

func AddFileToS3(fileDir string) string {

	creds := credentials.NewStaticCredentials(Settings.AwsAccessKeyId, Settings.AwsSecretAccessKey, "")
	_, err := creds.Get()
	if err != nil {
		log.Println("Error After set the creds")
	}

	// Create a single AWS session (we can re use this if we're uploading many files)
    cfg := aws.NewConfig().WithRegion(Settings.AwsS3RegionName).WithCredentials(creds)
	svc := s3.New(session.New(), cfg)
    // Upload

    // Open the file for use
    file, err := os.Open(fileDir)
    if err != nil {
        log.Println("Error After Open The File " + fileDir)
    }
    defer file.Close()

    // Get file size and read the file content into a buffer
    fileInfo, _ := file.Stat()
    var size int64 = fileInfo.Size()
    buffer := make([]byte, size)
	file.Read(buffer)
	fileName := filepath.Base(fileDir)

    // Config settings: this is where you choose the bucket, filename, content-type etc.
    // of the file you're uploading.
    resp, err := svc.PutObject(&s3.PutObjectInput{
        Bucket:               aws.String(Settings.AwsStorageBucketName),
        Key:                  aws.String(Settings.AwsS3Dir + fileName),
        Body:                 bytes.NewReader(buffer),
        ContentLength:        aws.Int64(size),
        ContentType:          aws.String(http.DetectContentType(buffer)),
	})

	if err != nil {
		log.Println("Error After Put Object to S3")
	}

	log.Println(awsutil.StringValue(resp))
	return Settings.AwsS3EndpointUrl + Settings.AwsS3Dir + fileName
}

func (c *Coral) getCoral(filename string) *Coral {
    yamlFile, err := ioutil.ReadFile( Settings.CoralDir + "/" + filename + ".yml")
    if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c);
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }

    return c
}

func readScheduleFiles() bool{
	content, err := ioutil.ReadFile(Settings.BuildDir + "/schedules.json")
    if err != nil {
        log.Fatal(err)
    }

	jsonErr := json.Unmarshal(content, &Schedules)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return true

}

func readBuildCommandsFiles() bool{
	content, err := ioutil.ReadFile(Settings.BuildDir + "/commands.json")
    if err != nil {
        os.OpenFile(Settings.BuildDir + "/commands.json", os.O_RDONLY|os.O_CREATE, 0755)
        log.Fatal(err)
    }

	jsonErr := json.Unmarshal(content, &BuildCommands)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return true

}

func readGreetingsFile() bool{
	content, err := ioutil.ReadFile(Settings.BuildDir + "/greetings.json")
    if err != nil {
        os.OpenFile(Settings.BuildDir + "/greetings.json", os.O_RDONLY|os.O_CREATE, 0755)
        log.Fatal(err)
    }

	jsonErr := json.Unmarshal(content, &BuildGreetings)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	return true

}


func fileSession(phone_number string) string {
    return Settings.BuildDir + "/sessions/" + phone_number + ".session"
}

// Simple helper function to read an environment or return a default value
func getEnvString(key string, defaultVal string) string {
    if value, exists := os.LookupEnv(key); exists {
	    return value
    }

    return defaultVal
}

// Simple helper function to read an environment or return a default value
func getEnvInt(key string, defaultVal int) int {
    valueStr := getEnvString(key, "")
    if value, err := strconv.Atoi(valueStr); err == nil {
	    return value
    }
    return defaultVal
}