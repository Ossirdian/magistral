# Payment Service

HTTP service for student payments and contract balances within the current `school_id`.

## Responsibilities

- CRUD payments
- store `amount`, `type`, `comment`
- store contract total per student
- calculate `paid_amount` and `remaining_amount`
- validate that `student_id` belongs to current school through `student` gRPC

## API

Contract lives in `shared/api/payments/v1`, generated package is `shared/pkg/payments/v1`.

Routes:

- `GET /payments`
- `POST /payments`
- `PATCH /payments/{paymentId}`
- `DELETE /payments/{paymentId}`
- `GET /payments/students/{studentId}/balance`
- `PUT /payments/students/{studentId}/contract`

## Notes

- `amount > 0`
- `contract_total >= 0`
- `remaining_amount = contract_total - paid_amount`
- current implementation allows overpayment, so `remaining_amount` may be negative
- `school_id` is taken from IAM validation context, not from request body
- student ownership is checked via `student` gRPC before create/balance/contract operations
- configure internal dependency with `STUDENT_GRPC_ADDR`
