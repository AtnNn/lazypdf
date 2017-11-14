#include <fitz.h>
#include <pdf/name-table.h>
#include <pdf/object.h>
#include <pdf/document.h>
#include <pthread.h>
#include <pdf/page.h>

// indent -linux -br -brf

// The number of mutexes we'll allocate
#define MUTEX_COUNT 10

fz_context *cgo_fz_new_context(const fz_alloc_context * alloc,
			       const fz_locks_context * locks,
			       size_t max_store);
int cgo_ptr_cast(ptrdiff_t ptr);
fz_document *cgo_open_document(fz_context *ctx, const char *filename);
void cgo_drop_document(fz_context *ctx, fz_document *doc);
void lock_mutex(void *locks, int lock_no);
void unlock_mutex(void *locks, int lock_no);
fz_locks_context *new_locks();
void free_locks(fz_locks_context * locks);
int get_rotation(fz_context *ctx, fz_page *page);
fz_buffer *getSVG(char *filename, int pageNum, fz_context *ctx);

typedef struct
{
	fz_document_writer super;
	char *path;
	int count;
	fz_output *out;
	int text_format;
	int reuse_images;
} fz_svg_writer;

struct fz_buffer_s
{
	int refs;
	unsigned char *data;
	size_t cap, len;
	int unused_bits;
	int shared;
};