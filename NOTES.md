# Notes

These are notes on the DDN project

## Technical todos
- [ ] Make submit button not tabindexable when hidden
- [ ] Implement flash (ROR term) messaging
- [ ] Show form validation errors other than root (which is already done)
- [ ] redirect urls ending in slash to the non slash counterparts
- [x] database pooling

## Business todos
- [ ] Search interface for inventory. Improvements
    - [ ] Should pull from `jobs/units` system once set up
    - [ ] Show "Showing x-y of z results" on the inventory page
    - [ ] Add color filter and other filters based on product attributes
- [ ] Jobs/units system
- [ ] Deletion improvements
    - [ ] Deletion confirmation; There should be an in-between page for deletion that makes a user confirm they actually want to at least for most things
    - [ ] Deleting with relations: Currently if you try to delete something that is referenced elsewhere it will fail with a "Something went wrong" Think about this process in more detail and fix it
- [ ] Email setup
    - [ ] Write email asking for sendgrid credentials including reason needed, amount needed to pay, how it will be used etc..
    - [ ] Setup sendgrid
    - [ ] Add tokens to dev and prod
    - [ ] Get client lib or manually requests if there isn't one setup in codebase
    - [ ] Implement proper password reset system
    - [ ] Implement welcome emails
- [ ]  User management improvements
    - [ ] Edit user action should work
    - [ ] Delete user action should work. Shouldn't be able to delete admins maybe? Maybe we need superadmins for this
- [ ] Product page polish
    - [ ] Implement multi colored buttons and make the different buttons different colors
    - [ ] Disable or hide the buttons that don't make sense based on what the state is

## Before initial launch
- [ ] Inventory deduction interface (precursor to jobs interface)
    - [ ] 2 sided interface
        - [ ] on left User flow
            - [x] They specify a qty in a text input
            - [x] Then they specify a product name 
            - [ ] Then they pick a bin location from a list filtered by the product name they picked
            - [ ] The list of bin locations is a styled checkboxes (looks like buttons). That will be in a wrapping flex container
            - [ ] The Deduct button will be disabled with a message saying you have a x out of y (better messsage later)
            - [ ] Once enough bin locations are checked to sum in available items to the qty or more that you specified then the btn isn't disabled
            - [ ] The minimum number of bins whose `qty` would sum to the `qty` that you specified or more will be selected
            - [ ] The btn would say "Add" when it is not disabled
            - [ ] After clicking it would move the data to the right
            - [ ] And clear all fields from the left
        - [ ] On the right
            - [ ] cumulative list of things to deduct is kept.
            - [ ] On each item
                - [ ] Display the:
                    - [ ] Qty
                    - [ ] Product Name
                    - [ ] bin location
                - [ ] There would be an edit button which moves the data to the left
                - [ ] There should be a remove btn
            - [ ] There is a btn to deduct
