package entinterceptor

// NewRelicSegmentDb new relic doc:
// https://entgo.io/docs/interceptors#examples
//func NewRelicSegmentDb() ent.Interceptor {
//	return ent.InterceptFunc(func(next ent.Querier) ent.Querier {
//		return ent.QuerierFunc(func(ctx context.Context, query ent.Query) (ent.Value, error) {
//
//			queryContext := entgo.QueryFromContext(ctx)
//			segment := helper.CreateNewRelicSegment(ctx, newrelic.DatastoreMySQL, queryContext.Type, queryContext.Op)
//			value, err := next.Query(ctx, query)
//
//			defer segment.End()
//			return value, err
//		})
//	})
//}
