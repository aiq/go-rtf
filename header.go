package rtf

type FontInfo struct {
	CharSet int
	Name    string
}

type FontTable map[int]FontInfo

type Color struct {
}

type ColorTable []Color

type Header struct {
	FontTable
	ColorTable
}
