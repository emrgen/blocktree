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

- [ ] create block
- [ ] update block
- [ ] delete block
- [ ] move block
- [ ] get block
- [ ] get all descendants of a block
- [ ] list blocks by parent
- [ ] list page blocks by space
- [ ] create space
- [ ] update space
- [ ] delete space
- [ ] get space
