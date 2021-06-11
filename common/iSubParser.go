package common

type ISubParser interface {

	DetermineFileTypeFromFile(filePath string) (*SubParserFileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*SubParserFileInfo, error)
}