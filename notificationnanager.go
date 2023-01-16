package notifications

import "context"

// Manager defines an interface for a notifications manager.
type Manager interface {
	// RegisterNotificationType registers a new type of notification.
	RegisterNotificationType(ctx context.Context)

	// GetNotificationTypes gets all the notification types.
	GetNotificationTypes(ctx context.Context)

	// SetKey sets a key under a specified namespace.
	// SetKey(ctx context.Context, key, namespace, value string) error
	// // GetKey returns the value for a combination of key and namespace, if set.
	// GetKey(ctx context.Context, key, namespace string) (string, error)
}
