package config

type Config struct {
	DocsDir      string
	DataDir      string
	OllamaURL    string
	OllamaModel  string
	Port         int
	MetadataFile string
	DBFile       string
	DevMode      bool
}
