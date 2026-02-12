package engine

import (
	"context"
	"testing"
)

func TestNoopAccessController(t *testing.T) {
	ctrl := &NoopAccessController{}
	ctx := context.Background()

	t.Run("CanAccessObject allows all", func(t *testing.T) {
		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("CanAccessObject should allow: %v", err)
		}

		err = ctrl.CanAccessObject(ctx, "AnyObject")
		if err != nil {
			t.Errorf("CanAccessObject should allow any object: %v", err)
		}
	})

	t.Run("CanAccessField allows all", func(t *testing.T) {
		err := ctrl.CanAccessField(ctx, "Account", "Name")
		if err != nil {
			t.Errorf("CanAccessField should allow: %v", err)
		}

		err = ctrl.CanAccessField(ctx, "AnyObject", "AnyField")
		if err != nil {
			t.Errorf("CanAccessField should allow any field: %v", err)
		}
	})
}

func TestDenyAllAccessController(t *testing.T) {
	ctrl := &DenyAllAccessController{}
	ctx := context.Background()

	t.Run("CanAccessObject denies all", func(t *testing.T) {
		err := ctrl.CanAccessObject(ctx, "Account")
		if err == nil {
			t.Error("CanAccessObject should deny")
		}
		if !IsAccessError(err) {
			t.Errorf("should return AccessError, got: %T", err)
		}

		accessErr := err.(*AccessError)
		if accessErr.Object != "Account" {
			t.Errorf("Object = %q, want %q", accessErr.Object, "Account")
		}
	})

	t.Run("CanAccessField denies all", func(t *testing.T) {
		err := ctrl.CanAccessField(ctx, "Account", "Name")
		if err == nil {
			t.Error("CanAccessField should deny")
		}
		if !IsAccessError(err) {
			t.Errorf("should return AccessError, got: %T", err)
		}

		accessErr := err.(*AccessError)
		if accessErr.Object != "Account" || accessErr.Field != "Name" {
			t.Errorf("AccessError = {%q, %q}, want {Account, Name}", accessErr.Object, accessErr.Field)
		}
	})
}

func TestObjectAccessController(t *testing.T) {
	ctx := context.Background()

	t.Run("nil AllowedObjects allows all", func(t *testing.T) {
		ctrl := &ObjectAccessController{AllowedObjects: nil}

		err := ctrl.CanAccessObject(ctx, "AnyObject")
		if err != nil {
			t.Errorf("nil AllowedObjects should allow all: %v", err)
		}
	})

	t.Run("empty AllowedObjects denies all", func(t *testing.T) {
		ctrl := &ObjectAccessController{AllowedObjects: map[string]bool{}}

		err := ctrl.CanAccessObject(ctx, "Account")
		if err == nil {
			t.Error("empty AllowedObjects should deny")
		}
	})

	t.Run("allows listed objects", func(t *testing.T) {
		ctrl := &ObjectAccessController{
			AllowedObjects: map[string]bool{
				"Account": true,
				"Contact": true,
			},
		}

		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("should allow Account: %v", err)
		}

		err = ctrl.CanAccessObject(ctx, "Contact")
		if err != nil {
			t.Errorf("should allow Contact: %v", err)
		}
	})

	t.Run("denies unlisted objects", func(t *testing.T) {
		ctrl := &ObjectAccessController{
			AllowedObjects: map[string]bool{
				"Account": true,
			},
		}

		err := ctrl.CanAccessObject(ctx, "Opportunity")
		if err == nil {
			t.Error("should deny Opportunity")
		}
	})

	t.Run("CanAccessField always allows", func(t *testing.T) {
		ctrl := &ObjectAccessController{
			AllowedObjects: map[string]bool{"Account": true},
		}

		err := ctrl.CanAccessField(ctx, "Account", "SSN")
		if err != nil {
			t.Errorf("should allow all fields: %v", err)
		}
	})
}

func TestFieldAccessController(t *testing.T) {
	ctx := context.Background()

	t.Run("nil maps allow all", func(t *testing.T) {
		ctrl := &FieldAccessController{}

		err := ctrl.CanAccessObject(ctx, "AnyObject")
		if err != nil {
			t.Errorf("nil AllowedObjects should allow: %v", err)
		}

		err = ctrl.CanAccessField(ctx, "AnyObject", "AnyField")
		if err != nil {
			t.Errorf("nil AllowedFields should allow: %v", err)
		}
	})

	t.Run("object access control", func(t *testing.T) {
		ctrl := &FieldAccessController{
			AllowedObjects: map[string]bool{
				"Account": true,
			},
		}

		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("should allow Account: %v", err)
		}

		err = ctrl.CanAccessObject(ctx, "Lead")
		if err == nil {
			t.Error("should deny Lead")
		}
	})

	t.Run("field access control", func(t *testing.T) {
		ctrl := &FieldAccessController{
			AllowedFields: map[string]map[string]bool{
				"Account": {
					"Name":     true,
					"Industry": true,
				},
			},
		}

		err := ctrl.CanAccessField(ctx, "Account", "Name")
		if err != nil {
			t.Errorf("should allow Name: %v", err)
		}

		err = ctrl.CanAccessField(ctx, "Account", "SSN")
		if err == nil {
			t.Error("should deny SSN")
		}
	})

	t.Run("no field restrictions for unlisted objects", func(t *testing.T) {
		ctrl := &FieldAccessController{
			AllowedFields: map[string]map[string]bool{
				"Account": {"Name": true},
			},
		}

		// Contact is not in AllowedFields, so all fields are allowed
		err := ctrl.CanAccessField(ctx, "Contact", "Email")
		if err != nil {
			t.Errorf("should allow all fields for unlisted object: %v", err)
		}
	})
}

func TestFuncAccessController(t *testing.T) {
	ctx := context.Background()

	t.Run("nil functions allow all", func(t *testing.T) {
		ctrl := &FuncAccessController{}

		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("nil ObjectFunc should allow: %v", err)
		}

		err = ctrl.CanAccessField(ctx, "Account", "Name")
		if err != nil {
			t.Errorf("nil FieldFunc should allow: %v", err)
		}
	})

	t.Run("custom object function", func(t *testing.T) {
		ctrl := &FuncAccessController{
			ObjectFunc: func(ctx context.Context, object string) error {
				if object == "Secret" {
					return NewAccessError(object)
				}
				return nil
			},
		}

		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("should allow Account: %v", err)
		}

		err = ctrl.CanAccessObject(ctx, "Secret")
		if err == nil {
			t.Error("should deny Secret")
		}
	})

	t.Run("custom field function", func(t *testing.T) {
		ctrl := &FuncAccessController{
			FieldFunc: func(ctx context.Context, object, field string) error {
				if field == "SSN" {
					return NewFieldAccessError(object, field)
				}
				return nil
			},
		}

		err := ctrl.CanAccessField(ctx, "Account", "Name")
		if err != nil {
			t.Errorf("should allow Name: %v", err)
		}

		err = ctrl.CanAccessField(ctx, "Account", "SSN")
		if err == nil {
			t.Error("should deny SSN")
		}
	})
}

func TestAccessControllerWithContext(t *testing.T) {
	type userIDKey struct{}

	ctrl := &FuncAccessController{
		ObjectFunc: func(ctx context.Context, object string) error {
			userID, ok := ctx.Value(userIDKey{}).(int)
			if !ok || userID == 0 {
				return NewAccessError(object)
			}
			return nil
		},
	}

	t.Run("without user context", func(t *testing.T) {
		ctx := context.Background()

		err := ctrl.CanAccessObject(ctx, "Account")
		if err == nil {
			t.Error("should deny without user")
		}
	})

	t.Run("with user context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userIDKey{}, 123)

		err := ctrl.CanAccessObject(ctx, "Account")
		if err != nil {
			t.Errorf("should allow with user: %v", err)
		}
	})
}

func TestAccessControllerInterfaceCompliance(t *testing.T) {
	// Compile-time interface compliance checks
	var _ AccessController = &NoopAccessController{}
	var _ AccessController = &DenyAllAccessController{}
	var _ AccessController = &ObjectAccessController{}
	var _ AccessController = &FieldAccessController{}
	var _ AccessController = &FuncAccessController{}
}
