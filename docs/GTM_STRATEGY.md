# Go-to-Market (GTM) Strategy & ICP

## 🎯 Ideal Customer Profile (ICP)
**The "Polyrepo" Power User**
- **Role**: Senior Engineers, SREs, Open Source Maintainers.
- **Pain Point**: "Context Fatigue" — wasting time checking status across 10+ microservices or libraries.
- **Solution**: A terminal-based "God Mode" dashboard for local git state.
- **Vibe**: Loves CLI tools, hates bloated GUIs (Electron apps), values speed and minimalism.

## 💎 Value Proposition
**Git-Scope** helps **busy developers** **instantly find, filter, and manage 50+ repos** so you can **stay in your flow state** without **endlessly `cd`-ing through nested directories.**

---

## 📢 Channel Strategy

### 1. Twitter / X (Build in Public)
**Goal**: Viral visual impression & peer validation.
**Strategy**: Post short, punchy videos (the GIF) + "Problem/Solution" threads.
**Tags**: `#golang` `#devtool` `#git` `#cli` `#opensource`

**Template 1 (Launch):**
> Stop `cd`-ing into 10 different folders just to check git status. 🛑
>
> I built **git-scope** — a blazing fast TUI to manage ALL your local repos in one view.
>
> 🚀 Features:
> • Recursive auto-discovery
> • Fuzzy search (`/`)
> • Vim navigation (`j`/`k`)
> • Instant editor jump (`Enter`)
>
> Written in Go + Bubble Tea 🍵.
>
> `brew install git-scope`
>
> [Link to GitHub]

**Template 2 (Philosophy):**
> GUIs are too slow. Manual `ls` is too tedious.
>
> The sweet spot is a TUI.
>
> Here's how I manage 20+ microservices without leaving the terminal:
> [GIF of Filtering Dirty Repos]

---

### 2. Reddit (Community Deep Dives)
**Goal**: Technical discussion and feedback.
**Subreddits**: `r/golang`, `r/commandline`, `r/git`, `r/devops`

**Title Ideas:**
- "I built a TUI to visualize 50+ git repos at once because I kept losing work"
- "Showcase: A Bubble Tea TUI for managing multi-repo workspaces"

**Body Structure:**
1. ** The Problem**: "I work on microservices and kept forgetting to push changes."
2. ** The Fix**: "Built a tool that caches filesystem state and gives a dashboard."
3. ** The Tech**: "Go, Bubble Tea, Lipgloss. 10ms startup time."
4. ** Call to Action**: "Brew installable. Feedback welcome!"

---

### 3. Hacker News (Show HN)
**Goal**: High-level exposure to engineering leaders.
**Timing**: Post on a weekday morning (8am - 10am PT).

**Title**: `Show HN: git-scope – Fast TUI to manage multiple git repositories`

**First Comment (Context):**
> "Hi HN, built this because `ls -R` wasn't cutting it for my local microservices setup. It scans your dev folder, caches the state, and lets you filter by 'dirty' status instantly. Written in Go."

---

### 4. Dev.to / Medium / Hashnode
**Goal**: SEO and long-tail traffic.
**Article Title**: "How to Tame Your Multi-Repo Chaos with the Terminal"
**Content**:
- Tutorial on setting up `git-scope`.
- Comparison with lazygit (lazygit is for *one* repo, git-scope is for *many* repos).

---

## 🗓️ Launch Checklist
- [ ] **Day 1**: Post "Show HN" + Reddit `r/golang`. Twitter announcement.
- [ ] **Day 2**: Respond to all comments. Push patches if bugs found.
- [ ] **Day 3**: Post "How I built this" thread on Twitter (technical breakdown).
- [ ] **Day 4**: Submit to `awesome-go` and `awesome-tui` lists.
