package filemeta

import "time"

type FileMeta struct {
	ID          string    `gorm:"type:varchar(40);primaryKey;column:id"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	FileName    *string   `gorm:"type:varchar(255);column:file_name"`
	FileSize    int64     `gorm:"column:file_size;not null"`
	FileKey     string    `gorm:"type:varchar(255);not null;column:file_key"`
	FileTxnType int16     `gorm:"column:file_txn_type;not null"`
	FileTxnMeta *string   `gorm:"type:varchar(255);column:file_txn_meta"`
	CompanyID   *string   `gorm:"type:varchar(40);column:company_id"`
}

func (FileMeta) TableName() string {
	return "files_meta"
}
