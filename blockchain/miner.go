package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
)

const (
	// MiningHeaderLen is the length of the MiningHeader struct in bytes
	MiningHeaderLen = (1 * (32 / 8)) + (3 * HashLen)
	// MaxUint32 is the max value for the uint32 type
	MaxUint32 uint32 = 4294967295
)

// MiningHeader contains the metadata required for mining
type MiningHeader struct {
	LastBlock Hash
	RootHash  Hash
	// Target is the current target stored in compact format
	Target Hash
	Nonce  uint32
}

// Marshal converts a Mining Header to a byte slice
func (mh *MiningHeader) Marshal() []byte {
	var buf []byte
	buf = append(buf, mh.LastBlock.Marshal()...)
	buf = append(buf, mh.RootHash.Marshal()...)
	buf = append(buf, mh.Target.Marshal()...)
	AppendUint32ToSlice(&buf, mh.Nonce)
	return buf
}

// SetMiningHeader sets the mining header
func (mh *MiningHeader) SetMiningHeader(lastBlock Hash, rootHash Hash, target Hash) {
	mh.LastBlock = lastBlock
	mh.RootHash = rootHash
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
		mh.Nonce++
	}
	return true
}

// VerifyMiningHeader confirms that the mining header is properly set
func (mh *MiningHeader) VerifyMiningHeader() bool {
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
