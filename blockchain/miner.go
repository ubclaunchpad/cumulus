package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"time"

	log "github.com/Sirupsen/logrus"
)

const (
	// MiningHeaderLen is the length of the MiningHeader struct in bytes
	MiningHeaderLen = (3 * HashLen) + (1 * (32 / 8)) + (1 * (64 / 8))
)

// MiningHeader contains the metadata required for mining
type MiningHeader struct {
	LastBlock Hash
	RootHash  Hash
	Target    Hash
	Time      uint32
	Nonce     uint64
}

// Marshal converts a Mining Header to a byte slice
func (mh *MiningHeader) Marshal() []byte {
	var buf []byte
	buf = append(buf, mh.LastBlock.Marshal()...)
	buf = append(buf, mh.RootHash.Marshal()...)
	buf = append(buf, mh.Target.Marshal()...)
	AppendUint32ToSlice(&buf, mh.Time)
	AppendUint64ToSlice(&buf, mh.Nonce)
	return buf
}

// SetMiningHeader sets the mining header
func (mh *MiningHeader) SetMiningHeader(lastBlock Hash, rootHash Hash, target Hash) {
	mh.LastBlock = lastBlock
	mh.RootHash = rootHash
	mh.Target = target
	mh.Time = uint32(time.Now().Unix())
	mh.Nonce = 0
}

// VerifyProofOfWork computes the hash of the MiningHeader and returns true if the result is less than the target
func (mh *MiningHeader) VerifyProofOfWork() bool {
	if CompareTo(mh.DoubleHashSum(), mh.Target, LessThan) {
		return true
	}

	return false
}

// DoubleHashSum computes the hash 256 of the marshalled mining header twice
func (mh *MiningHeader) DoubleHashSum() Hash {
	hash := sha256.Sum256(mh.Marshal())
	hash = sha256.Sum256(hash[:])
	return hash
}

// Mine continuously increases the nonce and tries to verify the proof of work until the puzzle is solved
func (mh *MiningHeader) Mine() bool {
	if mh.VerifyMiningHeader() == false {
		return false
	}

	for mh.VerifyProofOfWork() == false {
		if mh.Nonce == math.MaxUint64 {
			mh.Nonce = 0
		}
		mh.Time = uint32(time.Now().Unix())
		mh.Nonce++
	}
	return true
}

// VerifyMiningHeader confirms that the mining header is properly set
func (mh *MiningHeader) VerifyMiningHeader() bool {
	if mh.Time == 0 || mh.Time > uint32(time.Now().Unix()) {
		log.Error("Invalid time in mining header")
		return false
	}
	if CompareTo(mh.Target, MinHash, EqualTo) || CompareTo(mh.Target, MaxDifficulty, GreaterThan) {
		log.Error("Invalid target in mining header")
		return false
	}
	return true
}

// AppendUint32ToSlice converts a uint32 to a byte slice and appends it to a byte slice
func AppendUint32ToSlice(s *[]byte, num uint32) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, num)
	*s = append(*s, buf...)
}

// AppendUint64ToSlice converts a uint64 to a byte slice and appends it to a byte slice
func AppendUint64ToSlice(s *[]byte, num uint64) {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, num)
	*s = append(*s, buf...)
}
