package main


type Setting struct {
	UserAgent string `json:"USER_AGENT"`
	RandMin int `json:"RAND_MIN"`
	RandMax int `json:"RAND_MAX"`
	LimitRandMax int `json:"LIMIT_RAND_MAX"`
	SessionsDir string `json:"SESSION_DIR"`
	CoralDir string `json:"CORAL_DIR"`
	BuildDir string `json:"BUILD_DIR"`
	AwsS3RegionName string `json:"AWS_S3_REGION_NAME"`
	AwsS3Dir string `json:"AWS_S3_DIR"`
	AwsS3EndpointUrl string `json:"AWS_S3_ENDPOINT_URL"`
	AwsStorageBucketName string `json:"AWS_STORAGE_BUCKET_NAME"`
	AwsSecretAccessKey string `json:"AWS_SECRET_ACCESS_KEY"`
	AwsAccessKeyId string `json:"AWS_ACCESS_KEY_ID"`
}

type Commands struct {
	Prefix string `yaml:"prefix"`
	Command string `yaml:"command"`
	Message string `yaml:"message"`
	Record bool `yaml:"record"`
	RunProcess bool `yaml:"run_process"`
}

type ExitCommand struct {
	Prefix string `yaml:"prefix"`
	Command string `yaml:"command"`
	Message string `yaml:"message"`
}

type Validation struct {
	Rule string `yaml:"rule"`
	Message string `yaml:"message"`
}

type Question struct {
	Slug string `yaml:"slug"`
	Asking string `yaml:"asking"`
	Validation Validation `yaml:"validation"`
}

type Questions struct {
	Question Question
}

type Process struct {
	Timeout int `yaml:"timeout"`
	ExitCommand ExitCommand `yaml:"exit_command"`
	EndMessage string `yaml:"end_message"`
	Questions []Questions `yaml:"questions"`
}

type Author struct {
	Name string `yaml:"name"`
	Phone string `yaml:"phone"`
	Email string `yaml:"email"`
	Dept string `yaml:"dept"`
	BU string `yaml:"bussiness_unit"`
}

type Webhook struct {
	Service string `yaml:"service"` // SLACK, DISCORD, WEBHOOK
	URL string `yaml:"url"`
}

type Coral struct {
	Author Author `yaml:"author"`
	Schedule `yaml:"schedule"`
	DefaultGreeting Greeting `yaml:"default_greeting"`
	Commands Commands `yaml:"commands"`
	Process Process `yaml:"process"`
	Webhook Webhook `yaml:"webhook"`
	ExpectedUsers []string `yaml:"expected_users"`
}

type BuildCommand struct {
	Prefix string `json:"prefix"`
	Command string `json:"command"`
	Record bool `json:"record"`
	Message string `json:"message"`
	RunProcess string `json:"run_process"`
}

type pesanFetch struct {
	Message string `json:"message"`
}

type Data struct {
	Slug string `json:"slug"`
	Question string `json:"question"`
	Answer string `json:"answer"`
	Created string `json:"created"`
}

type Session struct {
	PhoneNumber string `json:"phone_number"`
	CurrentProcess string `json:"current_process"`
	CurrentQuestionSlug int `json:"current_question_slug"`
	ProcessStatus string `json:"process_status"` // DONE, WAIT_ANSWER, SENDED
	Datas []Data `json:"data"`
	Sent string `json:"sent"`
	SentTo string `json:"sent_to"`
	Created string `json:"created"`
	Expired string `json:"expired"`
	Finished string `json:"finished"`
}

type discord struct {
	Content string `json:"content"`
}

type Schedule struct {
	Rule string `json:"rule"`
	ProcessName string `json:"process_name"`
	Message string `json:"message" yaml:"message"`
	Sender string `json:"sender"`
	ExpectedUsers []string `json:"expected_users" yaml:"expected_users"`
}

type BuildGreeting struct {
	ProcessName string `json:"process_name"`
	Message string `json:"message" yaml:"message"`
	Webhook Webhook `json:"webhook" yaml:"webhook"`
	ExpectedUsers []string `json:"expected_users" yaml:"expected_users"`
}

type Greeting struct {
	Message string `yaml:"message"`
	Webhook Webhook `yaml:"webhook"`
}

type LogGreeting struct {
	Message string `json:"message" yaml:"message"`
	PhoneNumber string `json:"phone_number" yaml:"phone_number"`
}


// func NewProcess() {
// 	return &Process{
// 		Timeout: 300,
// 	}
// }