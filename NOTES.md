# Notes

These are notes on the DDN project

## Technical todos
- [ ] support 404s
- [ ] Make submit button not tabindexable when hidden
- [ ] Implement flash (ROR term) messaging
- [ ] Show form validation errors other than root (which is already done)
- [ ] redirect urls ending in slash to the non slash counterparts
- [x] database pooling
- [x] add `external_id` to `products` (example `W0930` W=it's a wall, 09 = width, 30 = height)
- [ ] Signals li'l lib
    - [ ] Make `<For>` component
    - [ ] Remove subscriptions to signals when they can be removed
    - [ ] Potentially make `effect` batch updates
    - [ ] Separate reading and writing in signals to allow for readonly state to be passed around (especially important for derivations)
- [ ] Ability to send email

## Business todos
- [ ] Search interface for inventory.
    - [x] Should pull from `inventory`
    - [ ] Should be filterable based on lots of different criteria
    - [ ] Should pull from `jobs/units` system once set up
- [ ] Jobs/units system
- [ ] Auth details
    - [ ] Sign up page isn't like what it really will be. In reaility admin users should be able to invite people
    - [ ] Password reset
    - [ ] Email verification before account creation
