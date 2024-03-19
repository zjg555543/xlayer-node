package pgpoolstorage

import (
	"context"
	"errors"

	"github.com/0xPolygonHermez/zkevm-node/pool"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jackc/pgx/v4"
)

// GetAllAddressesWhitelisted get all addresses whitelisted
func (p *PostgresPoolStorage) GetAllAddressesWhitelisted(ctx context.Context) ([]common.Address, error) {
	sql := `SELECT addr FROM pool.whitelisted`

	rows, err := p.db.Query(ctx, sql)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	var addrs []common.Address
	for rows.Next() {
		var addr string
		err := rows.Scan(&addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, common.HexToAddress(addr))
	}

	return addrs, nil
}

// BatchUpdateTxsStatus update tx status
func (p *PostgresPoolStorage) BatchUpdateTxsStatus(ctx context.Context, hashes []common.Hash, newStatus pool.TxStatus,
	isWIP bool, failedReason *string) error {
	sql := "UPDATE pool.transaction SET status = $1, is_wip = $2"

	hh := make([]string, 0, len(hashes))
	for _, h := range hashes {
		hh = append(hh, h.Hex())
	}

	if failedReason != nil {
		sql += ", failed_reason = $3 WHERE hash = ANY ($4)"

		if _, err := p.db.Exec(ctx, sql, newStatus, isWIP, failedReason, hh); err != nil {
			return err
		}
	} else {
		sql += " WHERE hash = ANY ($3)"

		if _, err := p.db.Exec(ctx, sql, newStatus, isWIP, hh); err != nil {
			return err
		}
	}

	return nil
}
