package plsql

import (
	"context"

	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/backend/plugin/parser/base"
	storepb "github.com/bytebase/bytebase/proto/generated-go/store"
)

func init() {
	base.RegisterGetQuerySpan(storepb.Engine_ORACLE, GetQuerySpan)
	base.RegisterGetQuerySpan(storepb.Engine_DM, GetQuerySpan)
	base.RegisterGetQuerySpan(storepb.Engine_OCEANBASE_ORACLE, GetQuerySpan)
}

func GetQuerySpan(ctx context.Context, statement string, database, schema string, getDatabaseMetadata base.GetDatabaseMetadataFunc, _ base.ListDatabaseNamesFunc, _ bool) (*base.QuerySpan, error) {
	extractor := newQuerySpanExtractor(database, schema, getDatabaseMetadata)

	querySpan, err := extractor.getQuerySpan(ctx, statement)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get query span from statement: %s", statement)
	}
	return querySpan, nil
}
