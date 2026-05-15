# Document Service

HTTP service for storing student document metadata, template-based generation, and file retrieval within the current `school_id`.

## Responsibilities

- CRUD document metadata
- upload document files into storage by `object_key`
- generate documents from templates
- link documents to students
- manage document templates
- validate that `student_id` belongs to current school through `student` gRPC

## API

Contract lives in `shared/api/documents/v1`, generated package is `shared/pkg/documents/v1`.

Routes:

- `GET /documents`
- `POST /documents/upload`
- `POST /documents/generate`
- `GET /documents/{documentId}`
- `GET /documents/{documentId}/download`
- `DELETE /documents/{documentId}`
- `GET /document-templates`
- `POST /document-templates`
- `PATCH /document-templates/{templateId}`
- `DELETE /document-templates/{templateId}`

## Notes

- metadata is stored in Postgres
- file bytes are stored outside DB under `STORAGE_ROOT`
- `school_id` is taken from IAM validation context, not from request body
- current local environment uses a filesystem-backed storage adapter behind the `object_key` abstraction
- student ownership is checked via `student` gRPC through `STUDENT_GRPC_ADDR`
