package lists

// Lists represent list data structure supported by the storage
//
//go:generate sh -c "rm -f mock/lists.go && mockgen -source=lists.go -destination=mock/lists.go -package=mocklists"
type Lists interface {
	LPush(key string, values ...[]byte) (int64, error)
}
