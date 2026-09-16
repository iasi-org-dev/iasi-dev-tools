IASI Dev depth/output refactor
==============================
- 4 spaces per indentation level (cli.IndentSize = 4)
- Step() remains one level; StepAt() accepts explicit depth
- Build/Publish/Release/Commit accept depth
- NothingToDo is silent
- targetUsesR() removed; lifecycle uses dispatchBuild/dispatchPublish/dispatchRelease
- website routed through the R-backed lifecycle
- operation messages include target name and type
- commit message is emitted only when there are actual changes
