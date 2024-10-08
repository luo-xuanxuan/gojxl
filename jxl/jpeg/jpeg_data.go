package jxl

import (
	"fmt"
	"gojxl/jxl"
	"gojxl/jxl/base"
)

const (
	kMaxComponents           int   = 4
	kMaxQuantTables          int   = 4
	kMaxHuffmanTables        int   = 4
	kJpegHuffmanMaxBitLength uint  = 16
	kJpegHuffmanAlphabetSize int   = 256
	kJpegDCAlphabetSize      int   = 12
	kMaxDHTMarkers           int   = 512
	kMaxDimPixels            int   = 65535
	kApp1                    uint8 = 0xE1
	kApp2                    uint8 = 0xE2
)

var (
	kIccProfileTag = []byte("ICC_PROFILE")
	kExifTag       = []byte("Exif\000")
	kXMPTag        = []byte("http://ns.adobe.com/xap/1.0/")

	kJPEGNaturalOrder [80]uint32 = [80]uint32{
		0, 1, 8, 16, 9, 2, 3, 10,
		17, 24, 32, 25, 18, 11, 4, 5,
		12, 19, 26, 33, 40, 48, 41, 34,
		27, 20, 13, 6, 7, 14, 21, 28,
		35, 42, 49, 56, 57, 50, 43, 36,
		29, 22, 15, 23, 30, 37, 44, 51,
		58, 59, 52, 45, 38, 31, 39, 46,
		53, 60, 61, 54, 47, 55, 62, 63,
		// extra entries for safety in decoder
		63, 63, 63, 63, 63, 63, 63, 63,
		63, 63, 63, 63, 63, 63, 63, 63,
	}

	kJPEGZigZagOrder [64]uint32 = [64]uint32{
		0, 1, 5, 6, 14, 15, 27, 28,
		2, 4, 7, 13, 16, 26, 29, 42,
		3, 8, 12, 17, 25, 30, 41, 43,
		9, 11, 18, 24, 31, 40, 44, 53,
		10, 19, 23, 32, 39, 45, 52, 54,
		20, 22, 33, 38, 46, 51, 55, 60,
		21, 34, 37, 47, 50, 56, 59, 61,
		35, 36, 48, 49, 57, 58, 62, 63,
	}
)

type JPEGQuantizationTable struct {
	values    [jxl.DCTBlockSize]int32
	precision uint32
	// The index of this quantization table as it was parsed from the input JPEG.
	// Each DQT marker segment contains an 'index' field, and we save this index
	// here. Valid values are 0 to 3.
	index uint32
	// Set to true if this table is the last one within its marker segment.
	isLast bool
}

func NewJPEGQuantizationTable() JPEGQuantizationTable {
	return JPEGQuantizationTable{
		isLast: true,
	}
}

type JPEGHuffmanCode struct {
	// Bit length histogram.
	counts [kJpegHuffmanMaxBitLength + 1]uint32
	// Symbol values sorted by increasing bit lengths.
	values [kJpegHuffmanAlphabetSize + 1]uint32
	// The index of the Huffman code in the current set of Huffman codes. For AC
	// component Huffman codes, 0x10 is added to the index.
	slotId int
	// Set to true if this Huffman code is the last one within its marker segment.
	isLast bool
}

func NewJPEGHuffmanCode() JPEGHuffmanCode {
	return JPEGHuffmanCode{
		isLast: true,
	}
}

// Huffman table indexes used for one component of one scan.
type JPEGComponentScanInfo struct {
	componentIndex uint32
	dcTableIndex   uint32
	acTableIndex   uint32
}

type ExtraZeroRunInfo struct {
	blockIndex       uint32
	numExtraZeroRuns uint32
}

// Contains information that is used in one scan.
type JPEGScanInfo struct {
	// Parameters used for progressive scans (named the same way as in the spec):
	//   Ss : Start of spectral band in zig-zag sequence.
	//   Se : End of spectral band in zig-zag sequence.
	//   Ah : Successive approximation bit position, high.
	//   Al : Successive approximation bit position, low.
	Ss            uint32
	Se            uint32
	Ah            uint32
	Al            uint32
	numComponents uint32
	components    [4]JPEGComponentScanInfo
	// Last codestream pass that is needed to write this scan.
	lastNeededPass uint32

	// Extra information required for bit-precise JPEG file reconstruction.

	// Set of block indexes where the JPEG encoder has to flush the end-of-block
	// runs and refinement bits.
	resetPoints []uint32

	// The number of extra zero runs (Huffman symbol 0xf0) before the end of
	// block (if nonzero), indexed by block index.
	// All of these symbols can be omitted without changing the pixel values, but
	// some jpeg encoders put these at the end of blocks.
	extraZeroRuns []ExtraZeroRunInfo
}

type coeff int16

type JPEGComponent struct {
	//four byte component id, might only need 1 byte?
	id uint32
	// Horizontal and vertical sampling factors.
	// In interleaved mode, each minimal coded unit (MCU) has
	// hSampleFactor x vSampleFactor DCT blocks from this component.
	hSampleFactor int
	vSampleFactor int
	// The index of the quantization table used for this component.
	quantizationIndex uint32
	// The dimensions of the component measured in 8x8 blocks.
	widthInBlocks  uint32
	heightInBlocks uint32
	// The DCT coefficients of this component, laid out block-by-block, divided
	// through the quantization matrix values.
	coeffs []coeff
}

func NewJPEGComponent() JPEGComponent {
	return JPEGComponent{
		hSampleFactor: 1,
		vSampleFactor: 1,
	}
}

type AppMarkerType uint32

const (
	kUnknown AppMarkerType = iota
	kICC
	kExif
	kXMP
)

type JPEGData struct {
	width           int
	height          int
	restartInterval uint32
	appData         [][]uint8
	appMarkerType   []AppMarkerType
	componentData   [][]uint8
	quantization    []JPEGQuantizationTable
	huffmanCode     []JPEGHuffmanCode
	components      []JPEGComponent
	scanInfo        []JPEGScanInfo
	markerOrder     []uint8
	interMarkerData [][]uint8
	tailData        []uint8

	hasZeroPaddingBit bool
	paddingBits       []uint8
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Mimics libjxl implementation, but we should be returning the values instead of modifying via pointers (maybe theres a perf dif?)
func (data *JPEGData) CalculateMCUSize(scan JPEGScanInfo, MCUsPerRow *int, MCURows *int) {
	isInterleaved := scan.numComponents > 1
	baseComponent := &data.components[scan.components[0].componentIndex]

	// hGroup / vGroup act as numerators for converting number of blocks to
	// number of MCU. In interleaved mode it is 1, so MCU is represented with
	// max*SampleFactor blocks. In non-interleaved mode we choose numerator to
	// be the samping factor, consequently MCU is always represented with single
	// block.
	hGroup := baseComponent.hSampleFactor
	vGroup := baseComponent.vSampleFactor
	if isInterleaved {
		hGroup = 1
		vGroup = 1
	}
	maxHSampleFactor := 1
	maxVSampleFactor := 1
	for _, component := range data.components {
		maxHSampleFactor = max(component.hSampleFactor, maxHSampleFactor)
		maxVSampleFactor = max(component.vSampleFactor, maxVSampleFactor)
	}
	*MCUsPerRow = base.DivCeil(data.width*hGroup, 8*maxHSampleFactor)
	*MCURows = base.DivCeil(data.height*vGroup, 8*maxVSampleFactor)
}

// TODO: implement VisitFields
func (data *JPEGData) VisitFields(visitor *jxl.Visitor) error {
	return nil
}

func SetJPEGDataFromICC(icc []uint8, jpegData *JPEGData) error {
	var iccPos uint = 0
	for i := range len(jpegData.appData) {
		if jpegData.appMarkerType[i] != kICC {
			continue
		}
		var length uint = uint(len(jpegData.appData[i])) - 17 //minus 17? idk its part of libjxl implementation
		if iccPos+length > uint(len(icc)) {
			return fmt.Errorf("ICC length is less than APP markers: requested %d more bytes, %d available", length, uint(len(icc))-iccPos)
		}
		copy(jpegData.appData[0][17:], icc[iccPos:iccPos+length]) //still dunno what 17 means
		iccPos += length
	}
	if iccPos != uint(len(icc)) && iccPos != 0 {
		return fmt.Errorf("ICC length is more then APP markers")
	}
	return nil
}
