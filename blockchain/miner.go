package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
)

const (
	// MiningHeaderLen is the length of the MiningHeader struct in bytes
	MiningHeaderLen = (4 * (32 / 8)) + (3 * HashLen)
	// Version is the current version of the proof of work function
	Version = 1
	// MaxUint32 is the max value for the uint32 type
	MaxUint32 uint32 = 4294967295
)

// MiningHeader contains the metadata required for mining
type MiningHeader struct {
	Version   uint32
	LastBlock Hash
	RootHash  Hash
	// Time is the current time as seconds since 1970-01-01T00:00 UTC
	Time uint32
	// Target is the current target stored in compact format
	Target Hash
	Nonce  uint32
}

// Marshal converts a Mining Header to a byte slice
func (mh *MiningHeader) Marshal() []byte {
	var buf []byte
	AppendUint32ToSlice(&buf, mh.Version)
	buf = append(buf, mh.LastBlock.Marshal()...)
	buf = append(buf, mh.RootHash.Marshal()...)
	AppendUint32ToSlice(&buf, mh.Time)
	buf = append(buf, mh.Target.Marshal()...)
	AppendUint32ToSlice(&buf, mh.Nonce)
	return buf
}

// SetMiningHeader sets the mining header
func (mh *MiningHeader) SetMiningHeader(lastBlock Hash, rootHash Hash, target Hash) {
	mh.Version = Version
	mh.LastBlock = lastBlock
	mh.RootHash = rootHash
	mh.Time = uint32(time.Now().Unix())
	mh.Target = target
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
		if mh.Nonce == MaxUint32 {
			return false
		}
		mh.Time = uint32(time.Now().Unix())
		mh.Nonce++
	}
	return true
}

// VerifyMiningHeader confirms that the mining header is properly set
func (mh *MiningHeader) VerifyMiningHeader() bool {
	if mh.Version != Version {
		return false
	}

	if mh.Time == 0 || mh.Time > uint32(time.Now().Unix()) {
		return false
	}

	if CompareTo(mh.Target, MinHash, EqualTo) || CompareTo(mh.Target, MaxDifficulty, GreaterThan) {
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
