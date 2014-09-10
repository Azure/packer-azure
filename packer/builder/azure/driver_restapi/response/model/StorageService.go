package model


type StorageService struct {
	Url 					string
	StorageServiceKeys 		StorageServiceKeys
}

type StorageServiceKeys struct {
	Primary 		string
	Secondary 		string
}


