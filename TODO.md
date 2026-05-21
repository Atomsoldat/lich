For now, implement for git

Using in memory copies sounds neat, and the library supports that, we should do that

## Figure out now

- run `fetch --all`
  - we need a Remote thingy
  - We get a Remote thingy from a Repository
  - A Repository can be instantiated with a working dir an a git dir



## Figure Out Next

- the relevant branch to checkout must be determined at the unit of work level
- we have units of work for animation, and now units of work for git
  - should those be related, or perhaps the same?
  - a kustomize unit of work represents an application
  - a git unit of work represents a "feature/bugfix" that we want to bring into the application
  - there can be many features/bugfixes, but only one application they belong to
  - this suggests they should be separate
  - the git-UOW could contain a pointer to a kustomize-UOW
  - how do we find out which kustomize-UOW belongs to which git-UOW?
    - one way would be looping over all kustomize-UOW and checking which the files in the modified git-UOW belong to (each kustomize-UOW has an origin directory)
    - we should still preferentially use the kustomisation we have found as the point of reference, in case of weirdness, and only use other files as a fallback (note, that detecting the kustomization.yaml is independent of it being modified by a renovate commit; its existence is a hard requirement)
    - **if** we fall back to other files, we find the closest kustomiszation.yaml in the directory tree, going up as needed
    - checking only for the kustomisation is not enough, because other files might have been touched by renovate 

## Migrate over the functionality of our shellscripts
- Detect
- Powerword Kill
- Command

## Streamline merging many branches

- Display a big list
- here are all the branches that have outstanding changes
- Go through them sequentially (or maybe just highlight and mark them simultaneously)
  - Skip or Process?
  - If Process, copy the repo to a temp location
  - Operate on it in a Goroutine
    - If an error occurs, queue for error resolution / review
  - When the operation is finished, queue the committed changes from the temp dir for review
- When we are done with the big list, or whenever we feel like it, we go to review mode
  - We then get to see the diffs that each branch created
  - Approve / Rework / Abort
    - Rework means "Requeue for additional,  possibly manual operations"

## Incomplete implementation of go-git
Go-Git does not provide the entire functionality of git. They are working on this, but they say themselves that git is probably too large to fully implement for them. In some cases, it seems like we would need to fall back to some other git library, or maybe git itself. Think about what that means and which choices we have.

- [libgit](https://github.com/libgit2/libgit2) C implementation
- git binary installed in the environment
- that one abandoned go git library whose name i forgot
# Later

## Temporary working directories
- in a monorepo, only one unit of work can check out its branch at a time
- we can fix that by copying the repo to another location and operating on it there
- do a shallow checkout, to the extent needed


## Other stuff than kustomize
There is no reason not to allow free form execution of other commands
But certain commands, which are particularly suited to GitOps should be offered as predefined building blocks