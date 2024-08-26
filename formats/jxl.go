package formats

type coeff int16

type jpegComponent struct {
	//four byte component id, might only need 1 byte?
	id uint32
	// Horizontal and vertical sampling factors.
	// In interleaved mode, each minimal coded unit (MCU) has
	// hSampleFactor x vSampleFactor DCT blocks from this component.
	hSampleFactor int64
	vSampleFactor int64
	// The index of the quantization table used for this component.
	quantizationIndex uint32
	// The dimensions of the component measured in 8x8 blocks.
	widthInBlocks  uint32
	heightInBlocks uint32
	// The DCT coefficients of this component, laid out block-by-block, divided
	// through the quantization matrix values.
	coeffs []coeff
}

func newJPEGComponent() jpegComponent {
	return jpegComponent{
		hSampleFactor: 1,
		vSampleFactor: 1,
	}
}
