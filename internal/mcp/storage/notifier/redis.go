package notifier

//type RedisNotifier struct {
//	client *redis.Client
//	topic  string
//}
//
//func (r *RedisNotifier) Notify(ctx context.Context, updated *config.MCPConfig) error {
//	data, _ := json.Marshal(updated)
//	return r.client.Publish(ctx, r.topic, data).Err()
//}
//
//func (r *RedisNotifier) Watch(ctx context.Context) (<-chan *config.MCPConfig, error) {
//	ch := make(chan *config.MCPConfig, 1)
//
//	pubsub := r.client.Subscribe(ctx, r.topic)
//	go func() {
//		defer close(ch)
//		for msg := range pubsub.Channel() {
//			var cfg config.MCPConfig
//			if err := json.Unmarshal([]byte(msg.Payload), &cfg); err == nil {
//				ch <- &cfg
//			}
//		}
//	}()
//
//	return ch, nil
//}
