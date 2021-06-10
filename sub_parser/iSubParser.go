package sub_parser

type ISubParser interface {

	DetermineFileTypeFromFile(filePath string) (*SubFileInfo, error)

	DetermineFileTypeFromBytes(inBytes []byte, nowExt string) (*SubFileInfo, error)
}