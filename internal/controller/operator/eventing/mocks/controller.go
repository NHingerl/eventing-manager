package mocks

import (
	kctrlruntimereconcile "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Update it when the issue: https://github.com/vektra/mockery/issues/787#issuecomment-2296180438 is fixed.

// Controller is an autogenerated mock type for the Controller type.
type Controller = TypedController[kctrlruntimereconcile.Request]
