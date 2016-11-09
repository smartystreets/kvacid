package acidic

import "time"

type KeyMapProjection struct {
	Head      uint64           `json:"head"`
	Tail      uint64           `json:"tail"`
	Committed map[string]*Item `json:"map"`
	open      map[string]map[string]*Item
}

func NewKeyMapProjection() *KeyMapProjection {
	return &KeyMapProjection{}
}

func (this *KeyMapProjection) Apply(message interface{}) {
	switch message := message.(type) {

	case TransactionStartedEvent:
		this.applyTransactionStarted(message)

	case ItemStoredEvent:
		this.applyItemStored(message)

	case DeletingItemEvent:
		this.applyDeletingItem(message)

	case TransactionCommittedEvent:
		this.applyTransactionCommitted(message)

	case TransactionFailedEvent:
		this.applyTransactionFailed(message)

	case TransactionAbortedEvent:
		this.applyTransactionAborted(message)

	case ItemMergedEvent:
		this.applyItemMerged(message)
	}
}

func (this *KeyMapProjection) applyTransactionStarted(message TransactionStartedEvent) {
	this.open[message.TransactionID] = make(map[string]*Item, 16)
}

func (this *KeyMapProjection) applyItemStored(message ItemStoredEvent) {
	this.findItem(message.TransactionID, message.CanonicalKey).UpdateStored(message.Sequence, message.Key, message.Revision)
}
func (this *KeyMapProjection) applyDeletingItem(message DeletingItemEvent) {
	this.findItem(message.TransactionID, message.Key).UpdateDeleted(message.Sequence)
}
func (this *KeyMapProjection) findItem(transactionID, key string) *Item {
	items := this.open[transactionID]
	if items == nil {
		return nil
	}

	item := items[key]
	if item == nil {
		item = &Item{}
		items[key] = item
	}

	return item
}

func (this *KeyMapProjection) applyTransactionCommitted(message TransactionCommittedEvent) {
	// move items to the committed index, set the head index to the sequence of the commit
	// CRITICAL BUG: all transactions must associated with a given commit must happen together
}

func (this *KeyMapProjection) applyTransactionFailed(message TransactionFailedEvent) {
	delete(this.open, message.TransactionID)
}

func (this *KeyMapProjection) applyTransactionAborted(message TransactionAbortedEvent) {
	delete(this.open, message.TransactionID)
}

func (this *KeyMapProjection) applyItemMerged(message ItemMergedEvent) {

}

func (this *KeyMapProjection) applyCommitMerged() {
	// TODO
}

type Item struct {
	Sequence   uint64    `json:"-"`
	Commit     uint64    `json:"commit,omitempty"`
	Expiration time.Time `json:"expiration,omitempty"`
	Key        string    `json:"key,omitempty"`
	Revision   string    `json:"version,omitempty"`
	Deleted    bool      `json:"deleted,omitempty"`
}

func (this *Item) UpdateDeleted(sequence, commit uint64) {
	if this != nil && sequence >= this.Sequence {
		this.Sequence = sequence
		this.Key = ""
		this.Revision = ""
		this.Deleted = true
		this.Expiration = 0
	}
}

func (this *Item) UpdateStored(sequence, commit uint64, key, revision string) {
	if this != nil && sequence >= this.Sequence {
		this.Sequence = sequence
		this.Key = key
		this.Revision = revision
		this.Deleted = false
		this.Expiration = 0
	}
}
