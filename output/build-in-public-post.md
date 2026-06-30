**The first users of most software in five years won't be people. They'll be agents. Almost no one is testing for that yet — so I want to show you what it actually looks like when you do.**

Here's the shift nobody's pricing in: software used to be built for a human at a screen. Click, read, decide. Every test we ever wrote assumed that human — the browser, the viewport, the mouse. But more and more products now have a different first customer: a model, running in some harness on a machine you'll never see, reading your docs and your CLI output and deciding *on its own* how to use you.

That's not a UI change. It's a change in who the user is. And it quietly breaks the entire testing stack underneath us.

We build a product like this — it installs *into* agent runtimes (Claude Code, Codex, others) as a CLI plus skills. No dashboard. The end user is the agent. Which forced a realization that I think is coming for everyone: **if your software is used by agents, the only honest way to test it is with agents.** Your test matrix stops being "browsers × operating systems" and becomes "harnesses × models."

So here's the rig, because the concrete version is more useful than the manifesto.

**Every test run is a disposable Docker container with a real agent baked in.** Not a mock of the agent — the actual third-party CLIs our users run. One image, a build arg for which mind goes in: just Codex, or Claude + another, or all of them. Same product, different agents, different runtime layouts. If the install only lands cleanly for one harness, we find out before a user does.

**The harness contains zero copy of the product.** It installs the real binary through the real installer. The bug that matters isn't "does my code work" — it's "does this survive contact with a mind I didn't write."

**Then we hand the agent a prompt and read what it leaves behind.** No scripted clicks — there's no screen to click. We tell it, roughly: *set up a monitor for this fictional company, run the tool, don't ask follow-up questions, make reasonable assumptions.* It installs, configures itself, drives the product, and we assert on the artifacts it produces. An agent grades the agent product; we check the homework.

The failures are the tell. They're almost never crashes. They're a perfectly capable model **confidently doing the wrong reasonable thing** because your install left it one ambiguity. A human would've shrugged and clicked around it. An agent commits. You only ever see those by letting a real one loose in a clean room and watching what it does — which is exactly the class of bug that ships to production invisibly today, because nothing in a normal test suite is built to catch it.

This is where I think software is going, fast:

→ Docs become an API, because the agent reads them as instructions, not suggestions.
→ Every ambiguity in your setup flow becomes a branch a model will pick wrong with total confidence.
→ "Works on my machine" dies for real — the clean room isn't optional when the user has no hands to fix things.
→ And your CI grows a population of agents, on different models, all trying to use you and reporting back.

We treat that as the product's actual proving ground: container in, behavior out, the thing earning its keep by surviving a mind it didn't write — over and over, across every harness we can throw at it.

The web spent twenty years learning to test for humans. We're going to spend the next few learning to test for everything else.

More soon. 🧃
