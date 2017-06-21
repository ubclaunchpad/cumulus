package miner

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
)

const (
	// MiningHeaderLen is the length of the MiningHeader struct in bytes
	MiningHeaderLen = (3 * blockchain.HashLen) + (1 * (32 / 8)) + (1 * (64 / 8))
)

var (
	// MinDifficulty is the minimum difficulty
	MinDifficulty = new(big.Int).Sub(BigExp(2, 232), big.NewInt(1))
	// MaxDifficulty is the maximum difficulty value
	MaxDifficulty = big.NewInt(1)
)

// MiningHeader contains the metadata required for mining
type MiningHeader struct {
	LastBlock blockchain.Hash
	RootHash  blockchain.Hash
	Target    blockchain.Hash
	Time      uint32
	Nonce     uint64
}

// Marshal converts a Mining Header to a byte slice
func (mh *MiningHeader) Marshal() []byte {
	var buf []byte
	buf = append(buf, mh.LastBlock.Marshal()...)
	buf = append(buf, mh.RootHash.Marshal()...)
	buf = append(buf, mh.Target.Marshal()...)
	tempBufTime := make([]byte, 4)
	binary.LittleEndian.PutUint32(tempBufTime, mh.Time)
	buf = append(buf, tempBufTime...)
	tempBufNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(tempBufNonce, mh.Nonce)
	buf = append(buf, tempBufNonce...)
	return buf
}

// SetMiningHeader sets the mining header (sets the time to the current time)
func (mh *MiningHeader) SetMiningHeader(lastBlock blockchain.Hash, rootHash blockchain.Hash, target blockchain.Hash) *MiningHeader {
	mh.LastBlock = lastBlock
	mh.RootHash = rootHash
	mh.Target = target
	mh.Time = uint32(time.Now().Unix())
	mh.Nonce = 0
	return mh
}

// VerifyProofOfWork computes the hash of the MiningHeader and returns true if the result is less than the target
func (mh *MiningHeader) VerifyProofOfWork() bool {
	return mh.DoubleHashSum().LessThan(mh.Target)
}

// DoubleHashSum computes the hash 256 of the marshalled mining header twice
func (mh *MiningHeader) DoubleHashSum() blockchain.Hash {
	hash := sha256.Sum256(mh.Marshal())
	hash = sha256.Sum256(hash[:])
	return hash
}

// Mine continuously increases the nonce and tries to verify the proof of work until the puzzle is solved
func (mh *MiningHeader) Mine() bool {
	if !mh.VerifyMiningHeader() {
		return false
	}

	for !mh.VerifyProofOfWork() {
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
	target := blockchain.HashToBigInt(mh.Target)
	// Check if the target is less than the max difficulty, or if the target is greater than the min difficulty
	if target.Cmp(MinDifficulty) == 1 || target.Cmp(MaxDifficulty) == -1 {
		log.Error("Invalid target in mining header")
		return false
	}
	return true
}

// BigExp returns an big int pointer with the result set to base**exp, if y <= 0, the result is 1
func BigExp(base, exp int) *big.Int {
	return new(big.Int).Exp(big.NewInt(int64(base)), big.NewInt(int64(exp)), nil)
}
