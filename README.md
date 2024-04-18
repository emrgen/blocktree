# Blocktree

A reusable library for block storage

## Hierarchical Data Model
```text
    0000
    ├── space-1
    │   ├── block-1.1
    │   └── block-1.2
    └── space-2
        ├── block-2.1
        └── block-2.2
```

## Internal Invariants

- all spaces are direct children of the root space(for now)
- space is a block
- blocks in a space form a tree (no cycles, no multiple parents, only one block tree per space)

## Progress

- [x] insert block
- [x] update block
- [x] patch block
- [x] delete block
- [x] erased block
- [x] move block
- [x] get block
- [x] get all descendants of a block
- [x] list blocks by parent
- [x] list page blocks by space
- [x] create space
- [x] update space
- [x] delete space
- [ ] get space
- [ ] get backlinks of a block
