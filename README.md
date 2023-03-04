# JPCBF (Japance Crossword broute-force solver)

A program for solving "Japanese crosswords" puzzles (Drawing by numbers) by linear enumeration of all possible positions

The program was written for fun and learning the "Go" programming language. After the first launch of its first version, it became immediately clear that the solution of almost any, even the smallest crossword puzzle, would take months or years. As a result, a small improvement was made, adding a preliminary analysis of lines using a classical algorithm in solving such crossword puzzles, but its implementation is very simple. But anyway, this allowed the program to solve some simple crossword puzzles within some acceptable time.


#### Usage:

	jpcbf [arguments] [crossword.json]:

#### Arguments:

| Argument | Type of value | Description |
|----------|------------|-------------|
| -c | int | Sets nuber of concurrent iterators (default 4) |
| -df | bool | Disable generating of filtering matrix |
| -p | int | Sets the maximum number of CPUs that can be executing simultaneousl<br />If n < 1, it does not change the current  (default -1) |
| -s | uint | Sets number of iterations per concurent iterator job (default 5000000) |

#### Puzzle (crossword) file format:

Crossword files are simply JSON files containing arrays of values for the horizontal and vertical keys of the puzzle. In "rows" arrays, values are read from left to right, and in "columns" arrays, values are read from bottom to top

Example:
```json
{
    "rows":
    [
		[4],
		[6],
		[8],
		[10],
		[10],
		[10],
		[2, 2, 2],
		[2, 2, 2],
		[2, 6],
		[2, 6],
    ],
    "columns":
    [
		[7],
		[8],
		[5],
		[6],
		[10],
		[10],
		[2, 6],
		[2, 5],
		[8],
		[7]
    ]
}
```
