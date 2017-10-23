package gpkg

import (
	log "github.com/sirupsen/logrus"

	"encoding/binary"
	"math"
	"fmt"
)

// Byte ordering flags
const wkbXDR = 0	// Big Endian
const wkbNDR = 1	// Little Endian

type WKBGeometry interface {
	// Initializes the geometry and returns the number of bytes consumed
	Init(bytes []byte) int
	Type() uint32
}

func bytesToUint32(bytes []byte, byteOrder uint8) uint32 {
	if len(bytes) != 4 {
		err := fmt.Errorf("Need 4 bytes to convert to uint32, received %v", len(bytes))
		log.Fatal(err)
	}

	var bitConversion binary.ByteOrder
	if byteOrder == wkbXDR {
		bitConversion = binary.BigEndian
	} else if byteOrder == wkbNDR {
		bitConversion = binary.LittleEndian
	} else {
		err := fmt.Errorf("Invalid byte order flag leading WKBGeometry: %v", byteOrder)
		log.Fatal(err)
	}
	
	value := bitConversion.Uint32(bytes)
	return value
}

func bytesToFloat64(bytes []byte, byteOrder uint8) float64 {
	if len(bytes) != 8 {
		err := fmt.Errorf("bytesToFloat64(): Need 8 bytes, received %v", len(bytes))
		log.Fatal(err)
	}

	var bitConversion binary.ByteOrder
	if byteOrder == wkbXDR {
		bitConversion = binary.BigEndian
	} else if byteOrder == wkbNDR {
		bitConversion = binary.LittleEndian
	} else {
		err := fmt.Errorf("Invalid byte order flag leading WKBGeometry: %v", byteOrder)
		log.Fatal(err)
	}

	bits := bitConversion.Uint64(bytes[:])
	value := math.Float64frombits(bits)
	
	return value	
}

type WKBPoint struct {
	x float64
	y float64
}

func (p *WKBPoint) Init(bytes []byte, byteOrder uint8) int {
	// Returns the number of bytes consumed
	if len(bytes) != 16 {
		err := fmt.Errorf("WKBPoint.Init(): Need 16 bytes, received %v", len(bytes))
		log.Fatal(err)
	}

	p.x = bytesToFloat64(bytes[:8], byteOrder)
	p.y = bytesToFloat64(bytes[8:16], byteOrder)
	
	return 16
}

type WKBLinearRing struct {
	numPoints 	uint32
	points 		[]WKBPoint;
}

func (lr *WKBLinearRing) Init(bytes []byte, byteOrder uint8) int {
	// Returns the number of bytes consumed
	if len(bytes) < 4 {
		err := fmt.Errorf("WKBLinearRing.Init(): Need at least 4 bytes, received %v", len(bytes))
		log.Fatal(err)
	}
	// Current read position of bytes
	i := 0
	lr.numPoints = bytesToUint32(bytes[i:4], byteOrder)
	lr.points = make([]WKBPoint, lr.numPoints)
	i += 4
	
	for p := uint32(0); p < lr.numPoints; p++ {
		lr.points[p].Init(bytes[i:i+16], byteOrder)
		i+=16
	}
	return i
}


type WKBPolygon struct {
	byteOrder byte
	wkbType uint32
	numRings uint32
	rings []WKBLinearRing
}


func (p *WKBPolygon) Init(bytes []byte) int {
	// Returns the number of bytes consumed
	i := 0
	byteOrder := bytes[i]
	i += 1
	p.byteOrder = byteOrder

	wkbType := bytesToUint32(bytes[i:i+4], byteOrder)
	polygonType := WKBTypeFlags["WKBPolygon"]
	if wkbType != polygonType {
		err := fmt.Errorf("Expected WKBPolygon type flag %v, got %v", polygonType, wkbType)
		log.Fatal(err)
	}
	p.wkbType = wkbType
	i += 4

	numRings := bytesToUint32(bytes[i:i+4], byteOrder)
	p.rings = make([]WKBLinearRing, numRings)
	i += 4

	for j := uint32(0); j < numRings; j++ {
		bytesConsumed := p.rings[j].Init(bytes[i:], byteOrder)
		i += bytesConsumed
	}
	
	return i
}

func (p *WKBPolygon) Type() uint32 {
	return p.wkbType
}

// Map WKBGeometry flag for type to GoLang type
var WKBTypeFlags map[string]uint32 = map[string]uint32 {
	"Geometry": 0,
	"Point": 1,
	"LineString": 2,
	"WKBPolygon": 3,
	"MultiPoint": 4,
	"MultiLineString": 5,
	"MultiPolygon": 6,
	"GeometryCollection": 7,
}

func newWKBGeometry(geomType uint32) WKBGeometry {
	switch geomType {
		case 3:
			return new(WKBPolygon)
		default:
			err := fmt.Errorf("newWKBGeometry: Unimplemented or invalid geomType: %v", geomType)
			log.Error(err)
			return nil
	}
}

func readNextGeometry(bytes []byte) (WKBGeometry, int) {
	// Returns number of bytes consumed
	if len(bytes) == 0 {
		return nil, 0
	}

	byteOrder := bytes[0]
	geomType := bytesToUint32(bytes[1:5], byteOrder)
	

	newGeom := newWKBGeometry(geomType)
	if newGeom == nil {
		err := fmt.Errorf("newWKBGeometry() returned nil for geomType %v", geomType)
		log.Fatal(err) 
	}
	bytesConsumed := newGeom.Init(bytes)
	
	return newGeom, bytesConsumed
}

func readGeometries(bytes []byte) ([]WKBGeometry, int) {
	// Returns an array of WKBGeometry instances and the number of bytes consumed
	var geoms []WKBGeometry
	
	// Current read location of bytes
	i := 0
	byteCount := len(bytes)
	log.Info("Reading WKBGeometries from data of length ", len(bytes))
	for i < byteCount {
		geom, bytesConsumed := readNextGeometry(bytes[i:])
		i += bytesConsumed
		if geom != nil {
			geoms = append(geoms, geom)
		}
	}

	if i != byteCount {
		err := fmt.Errorf("Bytes consumed from reading geometry (%v) doesn't match data length (%v)", i, byteCount)
		log.Warn(err) 
	}
	
	return geoms, i
}

type GeoPackageBinaryHeader struct {
	// See: http://www.geopackage.org/spec/
	initialized bool
    magic uint16	// should be 0x4750 (18256)
	version uint8
	flags uint8
	flagsReady bool
	srs_id int32
	envelope []float64
	headerSize int	// total bytes in header
}

func bytesToInt32(bytes []byte, byteOrder uint8) int32 {
	if len(bytes) != 4 {
		err := fmt.Errorf("Expecting 4 bytes, got %v", len(bytes))
		log.Error(err)
		return -1
	}

	valueLittleEndian := int32(bytes[3])
	valueLittleEndian <<= 8
	valueLittleEndian |= int32(bytes[2])
	valueLittleEndian <<= 8
	valueLittleEndian |= int32(bytes[1])
	valueLittleEndian <<= 8
	valueLittleEndian |= int32(bytes[0])

	valueBigEndian := int32(bytes[0])
	valueBigEndian <<= 8
	valueBigEndian |= int32(bytes[1])
	valueBigEndian <<= 8
	valueBigEndian |= int32(bytes[2])
	valueBigEndian <<= 8
	valueBigEndian |= int32(bytes[3])

	var value int32
	if byteOrder == wkbNDR {
		value = valueLittleEndian
	} else if byteOrder == wkbXDR {
		value = valueBigEndian
	}

	return value
}

func (h *GeoPackageBinaryHeader) Init(geom []byte) {
	const littleEndian = 1
	const bigEndian = 0
	
	h.flags = geom[3]
	h.flagsReady = true

	headerByteOrder := h.flags & 0x01

	var magic uint16
	
	if headerByteOrder == littleEndian {
		magic = uint16(geom[0]); magic <<= 8
		magic |= uint16(geom[1])
	} else {
		magic = uint16(geom[1]); magic <<= 8
		magic |= uint16(geom[0])
	}
	h.magic = magic
	
	h.version = geom[2]

	h.srs_id = bytesToInt32(geom[4:8], wkbNDR)

	if (h.srs_id < 0 || h.srs_id > 9999) {
		newByteOrder := uint8(littleEndian)
		if newByteOrder == headerByteOrder { newByteOrder = bigEndian }
		
		log.Errorf("The decoded srs_id (%v) looks invalid decoded with byteOrder %v", h.srs_id, headerByteOrder)
		log.Errorf("Trying again with byteOrder %v", newByteOrder)
		h.srs_id = bytesToInt32(geom[4:8], newByteOrder)
		if (h.srs_id < 0 || h.srs_id > 9999) {
			log.Errorf("Still looks invalid (%v), setting it to 0", h.srs_id)
			h.srs_id = 0
		} else {
			log.Errorf("Looks valid (%v) with this encoding, but not the encoding detected.", h.srs_id)
		}
	}
	
	eType := h.EnvelopeType()
	hSize := 8
	float64size := 8
	switch eType {
		case 0:
			h.envelope = make([]float64, 0)
		case 1:
			h.envelope = make([]float64, 4)
			hSize += 4 * float64size
			h.envelope[0] = bytesToFloat64(geom[8:16], headerByteOrder)
			h.envelope[1] = bytesToFloat64(geom[16:24], headerByteOrder)
			h.envelope[2] = bytesToFloat64(geom[24:32], headerByteOrder)
			h.envelope[3] = bytesToFloat64(geom[32:40], headerByteOrder)
		case 2:
			h.envelope = make([]float64, 6)
			hSize += 6 * float64size
			h.envelope[0] = bytesToFloat64(geom[8:16], headerByteOrder)
			h.envelope[1] = bytesToFloat64(geom[16:24], headerByteOrder)
			h.envelope[2] = bytesToFloat64(geom[24:32], headerByteOrder)
			h.envelope[3] = bytesToFloat64(geom[32:40], headerByteOrder)
			h.envelope[4] = bytesToFloat64(geom[40:48], headerByteOrder)
			h.envelope[5] = bytesToFloat64(geom[48:56], headerByteOrder)
		case 3:
			h.envelope = make([]float64, 6)
			hSize += 6 * float64size
			h.envelope[0] = bytesToFloat64(geom[8:16], headerByteOrder)
			h.envelope[1] = bytesToFloat64(geom[16:24], headerByteOrder)
			h.envelope[2] = bytesToFloat64(geom[24:32], headerByteOrder)
			h.envelope[3] = bytesToFloat64(geom[32:40], headerByteOrder)
			h.envelope[4] = bytesToFloat64(geom[40:48], headerByteOrder)
			h.envelope[5] = bytesToFloat64(geom[48:56], headerByteOrder)
		case 4:
			h.envelope = make([]float64, 8)
			hSize += 8 * float64size
			h.envelope[0] = bytesToFloat64(geom[8:16], headerByteOrder)
			h.envelope[1] = bytesToFloat64(geom[16:24], headerByteOrder)
			h.envelope[2] = bytesToFloat64(geom[24:32], headerByteOrder)
			h.envelope[3] = bytesToFloat64(geom[32:40], headerByteOrder)
			h.envelope[4] = bytesToFloat64(geom[40:48], headerByteOrder)
			h.envelope[5] = bytesToFloat64(geom[48:56], headerByteOrder)
			h.envelope[6] = bytesToFloat64(geom[56:64], headerByteOrder)
			h.envelope[7] = bytesToFloat64(geom[64:72], headerByteOrder)
		default:
			log.Errorf("Invalid envelope type: %v", eType)
			h.envelope = make([]float64, 0)
	}

	h.headerSize = hSize	

	log.Debugf("GeoPackageBinaryHeader.Init() header size: %v, geom blob size: %v", hSize, len(geom))

	h.initialized = true
}

func (h *GeoPackageBinaryHeader) isInitialized(caller string) bool {
	if !h.initialized {
		log.Errorf("%v: GeoPackageBinaryHeader not initialized", caller)
		return false
	} else {
		return true
	}
}

func (h *GeoPackageBinaryHeader) Magic() uint16 {
	if h.isInitialized("Magic()") {
		return h.magic
	} else {
		return uint16(0)
	}
}

func (h *GeoPackageBinaryHeader) Version() uint8 {
	if h.isInitialized("Version()") {
		return h.version
	} else {
		return 0
	}
}

func (h *GeoPackageBinaryHeader) EnvelopeType() uint8 {
	/*  0: no envelope (space saving slower indexing option), 0 bytes
	    1: envelope is [minx, maxx, miny, maxy], 32 bytes
	    2: envelope is [minx, maxx, miny, maxy, minz, maxz], 48 bytes
	    3: envelope is [minx, maxx, miny, maxy, minm, maxm], 48 bytes
	    4: envelope is [minx, maxx, miny, maxy, minz, maxz, minm, maxm], 64 bytes
	    5-7: invalid
	*/
	var envelope uint8
	if h.flagsReady {
		envelope = (h.flags & 0xE) >> 1
	} else {
		log.Errorf("GeoPackageBinaryHeader.flags must be ready before calling this function")
		envelope = 0
	}	

	return envelope
}

func (h *GeoPackageBinaryHeader) SRSId() int32 {
	if h.isInitialized("SRSId()") {
		return h.srs_id
	} else {
		return -1
	}
}

func (h *GeoPackageBinaryHeader) Envelope() []float64 {
	if h.isInitialized("Envelope()") {
		return h.envelope
	} else {
		return nil
	}
}

func (h *GeoPackageBinaryHeader) Size() int {
	if h.isInitialized("Size()") {
		return h.headerSize
	} else {
		return -1
	}
}