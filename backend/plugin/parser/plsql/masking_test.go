package plsql

import (
	"testing"

	"github.com/stretchr/testify/require"

	storepb "github.com/bytebase/bytebase/proto/generated-go/store"

	"github.com/bytebase/bytebase/backend/plugin/db"
)

func TestPLSQLExtractSensitiveField(t *testing.T) {
	const (
		defaultSchema = "ROOT"
	)
	var (
		defaultDatabaseSchema = &db.SensitiveSchemaInfo{
			DatabaseList: []db.DatabaseSchema{
				{
					Name: defaultSchema,
					SchemaList: []db.SchemaSchema{
						{
							Name: defaultSchema,
							TableList: []db.TableSchema{
								{
									Name: "T",
									ColumnList: []db.ColumnInfo{
										{
											Name:         "A",
											MaskingLevel: storepb.MaskingLevel_FULL,
										},
										{
											Name:         "B",
											MaskingLevel: storepb.MaskingLevel_NONE,
										},
										{
											Name:         "C",
											MaskingLevel: storepb.MaskingLevel_NONE,
										},
										{
											Name:         "D",
											MaskingLevel: storepb.MaskingLevel_PARTIAL,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	)
	tests := []struct {
		statement  string
		schemaInfo *db.SensitiveSchemaInfo
		fieldList  []db.SensitiveField
	}{
		{
			// Test for Recursive Common Table Expression dependent closures.
			statement: `
				with t1(cc1, cc2, cc3, n) as (
					select a as c1, b as c2, c as c3, 1 as n from t
					union all
					select cc1 * cc2, cc2 + cc1, cc3 * cc2, n + 1 from t1 where n < 5
				)
				select * from t1;
			`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "CC1",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "CC2",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "CC3",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "N",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
			},
		},
		{
			// Test for Recursive Common Table Expression.
			statement: `
				with t1 as (
					select 1 as c1, 2 as c2, 3 as c3, 1 as n from DUAL
					union all
					select c1 * a, c2 * b, c3 * d, n + 1 from t1, t where n < 5
				)
				select * from t1;
			`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "C1",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "C2",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C3",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
				{
					Name:         "N",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
			},
		},
		{
			// Test that Common Table Expression rename field names.
			statement:  `with t1(d, c, b, a) as (select * from t) select * from t1`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for Common Table Expression with UNION.
			statement:  `with t1 as (select * from t), t2 as (select * from t1) select * from (select * from t1 union all select * from t2)`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for Common Table Expression reference.
			statement:  `with t1 as (select * from t), t2 as (select * from t1) select * from t2`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for multi-level Common Table Expression.
			statement:  `with tt2 as (with tt2 as (select * from t) select MAX(A) from tt2) select * from tt2`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "MAX(A)",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
			},
		},
		{
			// Test for Common Table Expression.
			statement:  `with t1 as (select * from t) select * from t1`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for UNION.
			statement:  `select 1 as c1, 2 as c2, 3 as c3, 4 from DUAL UNION ALL select * from t`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "C1",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "C2",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C3",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "4",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for UNION.
			statement:  `select * from t UNION ALL select * from t`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for explicit schema name.
			statement:  `select CONCAT(ROOT.T.A, ROOT.T.B) from T`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "CONCAT(ROOT.T.A,ROOT.T.B)",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
			},
		},
		{
			// Test for associated sub-query.
			statement:  `select a, (SELECT MAX(B) > Y.A FROM T X) from t y`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "(SELECTMAX(B)>Y.AFROMTX)",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
			},
		},
		{
			// Test for JOIN with ON clause.
			statement:  `select * from t t1 join t t2 on t1.a = t2.a`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for natural JOIN.
			statement:  `select * from t t1 natural join t t2`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for JOIN with USING clause.
			statement:  `select * from t t1 join t t2 using(a)`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for non-associated sub-query
			statement:  "select t.a, (SELECT MAX(A) FROM T) from t t1 join t on t.a = t1.b",
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "(SELECTMAX(A)FROMT)",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
			},
		},
		{
			// Test for functions.
			statement:  `select A-B, B+C as c1 from (select * from t)`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A-B",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "C1",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
			},
		},
		{
			// Test for functions.
			statement:  `select MAX(A), min(b) as c1 from (select * from t)`,
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "MAX(A)",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "C1",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
			},
		},
		{
			// Test for sub-query
			statement:  "select * from (select * from t) where rownum <= 100000;",
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for sub-select.
			statement:  "select * from (select a, t.b, root.t.c, d as d1 from root.t) where ROWNUM <= 100000;",
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D1",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for field name.
			statement:  "select a, t.b, root.t.c, d as d1 from t",
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D1",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			statement:  "SELECT * FROM ROOT.T;",
			schemaInfo: defaultDatabaseSchema,
			fieldList: []db.SensitiveField{
				{
					Name:         "A",
					MaskingLevel: storepb.MaskingLevel_FULL,
				},
				{
					Name:         "B",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "C",
					MaskingLevel: storepb.MaskingLevel_NONE,
				},
				{
					Name:         "D",
					MaskingLevel: storepb.MaskingLevel_PARTIAL,
				},
			},
		},
		{
			// Test for EXPLAIN statements.
			statement:  "explain plan for select 1 from dual;",
			schemaInfo: &db.SensitiveSchemaInfo{},
			fieldList:  nil,
		},
		{
			// Test for no FROM DUAL.
			statement:  "select 1 from dual;",
			schemaInfo: &db.SensitiveSchemaInfo{},
			fieldList:  []db.SensitiveField{{Name: "1", MaskingLevel: storepb.MaskingLevel_NONE}},
		},
	}

	for _, test := range tests {
		extractor := &SensitiveFieldExtractor{
			CurrentDatabase: defaultSchema,
			SchemaInfo:      test.schemaInfo,
		}
		res, err := extractor.ExtractSensitiveField(test.statement)
		require.NoError(t, err, test.statement)
		require.Equal(t, test.fieldList, res, test.statement)
	}
}
