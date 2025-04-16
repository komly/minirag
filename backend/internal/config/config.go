package config

type Config struct {
	DocsDir          string
	DataDir          string
	OllamaURL        string
	OllamaModel      string
	OllamaEmbedModel string
	Port             int
	MetadataFile     string
	DBFile           string
	DevMode          bool
	ForceReindex     bool
}
