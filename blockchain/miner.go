package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"strconv"
	"time"
)

const (
	// MiningHeaderLen is the length of the MiningHeader struct in bytes
	MiningHeaderLen = (4 * (32 / 8)) + (2 * HashLen)
	// CurrentVersion is the current version of the proof of work function
	CurrentVersion = 1
	// MaxUint32 is the max value for the uint32 type
	MaxUint32 = 4294967295
)

// MiningHeader contains the metadata required for mining
type MiningHeader struct {
	Version   uint32
	LastBlock Hash
	RootHash  Hash
	Time      uint32
	Target    uint32
	Nonce     uint32
}

// Marshal converts a Mining Header to a byte slice
func (mh *MiningHeader) Marshal() []byte {
	var buf []byte
	AppendUint32ToSlice(&buf, mh.Version)
	buf = append(buf, mh.LastBlock.Marshal()...)
	buf = append(buf, mh.RootHash.Marshal()...)
	AppendUint32ToSlice(&buf, mh.Time)
	AppendUint32ToSlice(&buf, mh.Target)
	AppendUint32ToSlice(&buf, mh.Nonce)
	return buf
}

// SetMiningHeader sets the mining header
func (mh *MiningHeader) SetMiningHeader(lastBlock Hash, rootHash Hash, target uint32) {
	rawTimeString := time.Now().String()
	formattedTimeString := rawTimeString[2:4] + rawTimeString[5:7] + rawTimeString[8:10] + rawTimeString[11:13] + rawTimeString[14:16]
	formattedTime, e := strconv.ParseUint(formattedTimeString, 10, 32)

	if e != nil {
		log.Fatal(e)
	}

	time := uint32(formattedTime)

	mh.Version = CurrentVersion
	// TODO: have the block hashes be pulled from blockchain
	mh.LastBlock = lastBlock
	mh.RootHash = rootHash
	// TODO: have the current time be in unix time
	mh.Time = uint32(time)
	mh.Target = target
	mh.Nonce = 0
}

// VerifyProofOfWork computes the hash of the MiningHeader and returns true if the result is less than the target
func VerifyProofOfWork(mh *MiningHeader) bool {
	hash := sha256.Sum256(mh.Marshal())
	hash = sha256.Sum256(hash[:])

	if CompareTo(hash, CompactToHash(mh.Target)) == LessThan {
		return true
	}

	return false
}

// Mine continuously increases the nonce and tries to verify the proof of work until the puzzle is solved
func Mine(mh *MiningHeader) bool {
	for VerifyProofOfWork(mh) == false {
		if mh.Nonce == MaxUint32 {
			return false
		}
		mh.Nonce++
	}
	return true
}

// AppendUint32ToSlice converts a uint32 to a byte slice and appends it to a byte slice
func AppendUint32ToSlice(s *[]byte, num uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	*s = append(*s, buf...)
}

// BaseConverter converts a decimal number to a number of a specified base, in hex
func BaseConverter(a uint32, b uint32) string {
	if a < b {
		s := strconv.FormatUint(uint64(a), 16)
		if len(s) == 1 {
			return "0" + s
		}
		return s
	}
	return BaseConverter(a/b, b) + strconv.FormatUint(uint64(a%b), 16)
}
