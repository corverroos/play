# Play

This is a `player` implementation of the [Unsure Tournament](https://github.com/corverroos/unsure).

## State Machine

The player implements a `internal.Round` state machine for each round it plays.
```
                +------------------------+
                |                        v
Joined +--> Collected --> Shared +--> Submitted 
       |                     ^   |
       L--> Excluded         +---+
```

The state of this player and all other players are also stored in the `internal.Round` entity in the `RoundState` unstructured json field.

## Consumers

All consumer logic is driven by reacting to events, either engine events or other player events.

The following *Engine Events* consumers are implemented:
- On RoundJoin event, call `engine.JoinRound` and insert an `internal.Round` entry in the `Joined` state.
- On RoundCollect event, if excluded shift to `Excluded` status, else call `engine.CollectRound` and shift to `Collected` updating rank and parts.
- On RoundSubmit event, if shared data with all other players and I am first, call `engine.SubmitRound` and shift to `Submitted` state.
- On MatchEnded event, exit the app.

The following *Play Events* consumers are implemented for each other player:
- On Collected or Excluded event, call `play.GetRoundData` and shift to `Shared` state updating inclusion, rank and parts of that player.
- On Submitted event, if shared data with all other players and I am next, call `engine.SubmitRound` and shift to `Submitted` state.

## Run

To run locally:

```
# Run the engine
git clone https://github.com/corverroos/unsure.git 
cd unsure
go run engine/engine/main.go --db_recreate --crash_ttl=0 --fate_p=0.02 --rounds=2

# Run three players
cd <to_this_repo>
go run play/main.go --db_recreate --engine_address="127.0.0.1:12048" --crash_ttl=0 --fate_p=0.02 -count=3 -index=0 &
go run play/main.go --db_recreate --engine_address="127.0.0.1:12048" --crash_ttl=0 --fate_p=0.02 -count=3 -index=1 &
go run play/main.go --db_recreate --engine_address="127.0.0.1:12048" --crash_ttl=0 --fate_p=0.02 -count=3 -index=2 &
```