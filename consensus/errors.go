package consensus

// TransactionCode is returned from ValidTransaction.
type TransactionCode uint32

// CloudBaseTransactionCode is returned from ValidCloudBaseTransaction.
type CloudBaseTransactionCode uint32

// GenesisBlockCode is returned from ValidGenesisBlock.
type GenesisBlockCode uint32

// BlockCode is returned from ValidBlock.
type BlockCode uint32

const (
	// ValidTransaction is returned when transaction is valid.
	ValidTransaction TransactionCode = iota
	// NoInputTransaction is returned when transaction has no valid input transaction.
	NoInputTransaction
	// Overspend is returned when transaction outputs exceed transaction inputs.
	Overspend
	// BadSig is returned when the signature verification fails.
	BadSig
	// Respend is returned when inputs have been spent elsewhere in the chain.
	Respend
	// NilTransaction is returned when the transaction pointer is nil.
	NilTransaction
)

const (
	// ValidCloudBaseTransaction is returned when a transaction is a valid
	// CloudBase transaction.
	ValidCloudBaseTransaction CloudBaseTransactionCode = iota
	// BadCloudBaseSender is returned when the sender address in the CloudBase
	// transaction is not a NilAddr.
	BadCloudBaseSender
	// BadCloudBaseInput is returned when all the fields inf the  CloudBase
	// transaction input are not equal to 0.
	BadCloudBaseInput
	// BadCloudBaseOutput is returned when the CloudBase transaction output is
	// invalid.
	BadCloudBaseOutput
	// BadCloudBaseReward is returned when the CloudBase transaction reward is
	// invalid.
	BadCloudBaseReward
	// BadCloudBaseSig is returned when the CloudBase transaction signature is
	// not equal to NilSig.
	BadCloudBaseSig
	// NilCloudBaseTransaction is returned when the CloudBase transaction
	// pointer is nil
	NilCloudBaseTransaction
)

const (
	// ValidGenesisBlock is returned when a block is a valid genesis block.
	ValidGenesisBlock GenesisBlockCode = iota
	// BadGenesisLastBlock is returned when the LastBlock of the genesis block
	// is not equal to 0.
	BadGenesisLastBlock
	// BadGenesisTransactions is returned when the genesis block does not contain
	// exactly 1 transaction, the first CloudBase transaction.
	BadGenesisTransactions
	// BadGenesisCloudBaseTransaction is returned when the transaction in the
	// genesis block is not a valid CloudBase transaction.
	BadGenesisCloudBaseTransaction
	// BadGenesisBlockNumber is returned when the block number in the genesis
	// block is not equal to 0.
	BadGenesisBlockNumber
	// BadGenesisTarget is returned when the genesis block's target is invalid.
	BadGenesisTarget
	// BadGenesisTime is returned when the genesis block's time is invalid.
	BadGenesisTime
	// NilGenesisBlock is returned when the genesis block is equal to nil.
	NilGenesisBlock
)

const (
	// ValidBlock is returned when the block is valid.
	ValidBlock BlockCode = iota
	// BadTransaction is returned when the block contains an invalid
	// transaction.
	BadTransaction
	// BadTime is returned when the block contains an invalid time.
	BadTime
	// BadTarget is returned when the block contains an invalid target.
	BadTarget
	// BadBlockNumber is returned when block number is not one greater than
	// previous block.
	BadBlockNumber
	// BadHash is returned when the block contains incorrect hash.
	BadHash
	// DoubleSpend is returned when two transactions in the block share inputs,
	// but outputs > inputs.
	DoubleSpend
	// BadCloudBaseTransaction is returned when a block does not have a
	// CloudBase transaction as the first transaction in its list of
	// transactions.
	BadCloudBaseTransaction
	// BadGenesisBlock is returned if the block is a genesis block and is
	// invalid.
	BadGenesisBlock
	// BadNonce is returned if the nonce is invalid.
	BadNonce
	// NilBlock is returned when the block pointer is nil.
	NilBlock
)
