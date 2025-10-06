package main

import (
	"context"
	"sync"
	"time"
)

// PersistentMinimaxBot represents a bot that maintains a persistent search tree
// and continues calculating during opponent's thinking time
type PersistentMinimaxBot struct {
	Symbol       byte
	Name         string
	InitialDepth int
	Base         int

	// Tree management
	rootNode *SearchNode
	tree     *SearchTree
	mutex    sync.RWMutex
}

// SearchNode represents a node in the persistent search tree
type SearchNode struct {
	ID           string // unique identifier
	Board        *Board // game state at this node
	Move         string // move that led to this state (empty for root)
	Depth        int    // depth in the search tree
	Score        int    // minimax score
	IsMaximizing bool   // whether this is a maximizing node

	// Tree structure
	Parent   *SearchNode            // parent node
	Children map[string]*SearchNode // child nodes keyed by move

	// Goroutine management
	ctx       context.Context    // context for this node's goroutine
	cancel    context.CancelFunc // cancellation function
	goroutine chan struct{}      // signals when goroutine is running

	// Synchronization
	mutex       sync.RWMutex // protects node data
	expanded    bool         // whether children have been generated
	calculating bool         // whether currently calculating
}

// SearchTree manages the persistent search tree
type SearchTree struct {
	root     *SearchNode
	maxDepth int                    // current maximum search depth
	nodes    map[string]*SearchNode // all active nodes
	mutex    sync.RWMutex           // protects tree structure

	// Background calculation
	expandQueue chan *SearchNode   // nodes waiting to be expanded
	ctx         context.Context    // global context
	cancel      context.CancelFunc // global cancellation
	wg          sync.WaitGroup     // tracks active goroutines
}

// NewPersistentMinimaxBot creates a new persistent minimax bot
func NewPersistentMinimaxBot(symbol byte, name string, initialDepth int, base int) *PersistentMinimaxBot {
	bot := &PersistentMinimaxBot{
		Symbol:       symbol,
		Name:         name,
		InitialDepth: initialDepth,
		Base:         base,
	}

	// Initialize search tree
	ctx, cancel := context.WithCancel(context.Background())
	bot.tree = &SearchTree{
		maxDepth:    initialDepth,
		nodes:       make(map[string]*SearchNode),
		expandQueue: make(chan *SearchNode, 100), // buffered queue
		ctx:         ctx,
		cancel:      cancel,
	}

	// Start background worker for expanding nodes
	go bot.tree.backgroundExpander()

	return bot
}

// MakeMove implements BotInterface
func (bot *PersistentMinimaxBot) MakeMove(board *Board) (string, [3]int) {
	bot.mutex.Lock()
	defer bot.mutex.Unlock()

	// Initialize or update root node
	if bot.rootNode == nil {
		bot.initializeRoot(board)
	} else {
		// Update root based on current board state
		bot.updateRoot(board)
	}

	// Wait for initial search to complete or timeout
	bestMove := bot.getBestMove()

	// Debug: Check if we have valid moves
	if bestMove == "" {
		validMoves := board.GetValidMoves()
		if len(validMoves) > 0 {
			bestMove = validMoves[0] // Emergency fallback
		}
	}

	// Execute the move
	coords := [3]int{-1, -1, -1}
	if bestMove != "" {
		coords = board.Move(bestMove, bot.Symbol)

		// Update root to reflect our move
		bot.moveRoot(bestMove)
	}

	return bestMove, coords
}

// OpponentMove notifies the bot of opponent's move for tree pruning
func (bot *PersistentMinimaxBot) OpponentMove(move string) {
	bot.mutex.Lock()
	defer bot.mutex.Unlock()

	if bot.rootNode != nil {
		bot.moveRoot(move)
	}
}

// initializeRoot creates the initial root node and starts search
func (bot *PersistentMinimaxBot) initializeRoot(board *Board) {
	rootID := "root"
	ctx, cancel := context.WithCancel(bot.tree.ctx)

	bot.rootNode = &SearchNode{
		ID:           rootID,
		Board:        copyBoard(board),
		Move:         "",
		Depth:        0,
		IsMaximizing: bot.Symbol == 'x',
		Children:     make(map[string]*SearchNode),
		ctx:          ctx,
		cancel:       cancel,
		goroutine:    make(chan struct{}),
	}

	bot.tree.root = bot.rootNode
	bot.tree.nodes[rootID] = bot.rootNode

	// Start expanding from root
	go bot.expandNode(bot.rootNode)
}

// updateRoot updates the root to match current board state
func (bot *PersistentMinimaxBot) updateRoot(board *Board) {
	// For now, reinitialize if board state doesn't match
	// TODO: Implement smart root finding based on board comparison
	bot.cleanup()
	bot.initializeRoot(board)
}

// moveRoot shifts the root to a child node and prunes irrelevant branches
func (bot *PersistentMinimaxBot) moveRoot(move string) {
	if bot.rootNode == nil {
		return
	}

	bot.tree.mutex.Lock()
	defer bot.tree.mutex.Unlock()

	// Find the child corresponding to the move
	newRoot, exists := bot.rootNode.Children[move]
	if !exists {
		// Move not in our search tree, reinitialize
		bot.cleanup()
		return
	}

	// Kill all other branches
	for childMove, child := range bot.rootNode.Children {
		if childMove != move {
			bot.killBranch(child)
		}
	}

	// Update tree structure
	oldRoot := bot.rootNode
	bot.rootNode = newRoot
	bot.tree.root = newRoot
	newRoot.Parent = nil
	newRoot.Depth = 0

	// Update depths of all descendants
	bot.updateDepths(newRoot, 0)

	// Clean up old root
	oldRoot.cancel()
	delete(bot.tree.nodes, oldRoot.ID)
}

// getBestMove returns the best move from current search
func (bot *PersistentMinimaxBot) getBestMove() string {
	if bot.rootNode == nil {
		return ""
	}

	// Wait a bit for initial expansion
	time.Sleep(100 * time.Millisecond)

	bot.rootNode.mutex.RLock()

	// If no children yet, use valid moves directly
	if len(bot.rootNode.Children) == 0 {
		bot.rootNode.mutex.RUnlock()
		validMoves := bot.rootNode.Board.GetValidMoves()
		if len(validMoves) > 0 {
			return validMoves[0] // Return first valid move as fallback
		}
		return ""
	}

	bestMove := ""
	bestScore := MIN_INT
	if !bot.rootNode.IsMaximizing {
		bestScore = MAX_INT
	}

	for move, child := range bot.rootNode.Children {
		child.mutex.RLock()
		score := child.Score
		child.mutex.RUnlock()

		if bot.rootNode.IsMaximizing && score > bestScore {
			bestScore = score
			bestMove = move
		} else if !bot.rootNode.IsMaximizing && score < bestScore {
			bestScore = score
			bestMove = move
		}
	}

	bot.rootNode.mutex.RUnlock()

	// Fallback to first child if no best move found
	if bestMove == "" {
		bot.rootNode.mutex.RLock()
		for move := range bot.rootNode.Children {
			bestMove = move
			break
		}
		bot.rootNode.mutex.RUnlock()
	}

	return bestMove
}

// expandNode runs as a goroutine to expand a search node
func (bot *PersistentMinimaxBot) expandNode(node *SearchNode) {
	defer bot.tree.wg.Done()
	bot.tree.wg.Add(1)

	close(node.goroutine) // Signal that goroutine is running

	for {
		select {
		case <-node.ctx.Done():
			return // Node was cancelled

		default:
			node.mutex.Lock()

			// Check if we should expand (are we at current max depth?)
			if node.Depth >= bot.tree.maxDepth {
				// We're a leaf, calculate score if not done
				if !node.calculating {
					node.calculating = true
					node.Score = node.Board.Evaluate()
					bot.propagateScore(node)
				}
				node.mutex.Unlock()

				// Wait for depth increase or cancellation
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Generate children if not expanded
			if !node.expanded {
				validMoves := node.Board.GetValidMoves()
				symbol := byte('x')
				if !node.IsMaximizing {
					symbol = 'o'
				}

				for _, move := range validMoves {
					childBoard := copyBoard(node.Board)
					childBoard.Move(move, symbol)

					childID := node.ID + "_" + move
					ctx, cancel := context.WithCancel(node.ctx)

					child := &SearchNode{
						ID:           childID,
						Board:        childBoard,
						Move:         move,
						Depth:        node.Depth + 1,
						IsMaximizing: !node.IsMaximizing,
						Parent:       node,
						Children:     make(map[string]*SearchNode),
						ctx:          ctx,
						cancel:       cancel,
						goroutine:    make(chan struct{}),
						Score:        childBoard.Evaluate(), // Initialize with board evaluation
					}

					node.Children[move] = child
					bot.tree.nodes[childID] = child

					// Start goroutine for child
					go bot.expandNode(child)
				}

				node.expanded = true

				// Immediately propagate initial scores up
				bot.propagateScore(node)
			}

			node.mutex.Unlock()

			// Wait before next iteration
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// propagateScore propagates a score change up the tree
func (bot *PersistentMinimaxBot) propagateScore(node *SearchNode) {
	current := node.Parent

	for current != nil {
		current.mutex.Lock()

		// Recalculate score based on children
		if len(current.Children) > 0 {
			bestScore := MIN_INT
			if !current.IsMaximizing {
				bestScore = MAX_INT
			}

			for _, child := range current.Children {
				child.mutex.RLock()
				score := child.Score
				child.mutex.RUnlock()

				if current.IsMaximizing && score > bestScore {
					bestScore = score
				} else if !current.IsMaximizing && score < bestScore {
					bestScore = score
				}
			}

			current.Score = bestScore
		}

		current.mutex.Unlock()
		current = current.Parent
	}
}

// backgroundExpander runs background expansion of leaf nodes
func (tree *SearchTree) backgroundExpander() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tree.ctx.Done():
			return

		case <-ticker.C:
			// Periodically increase search depth
			tree.mutex.Lock()
			tree.maxDepth++
			tree.mutex.Unlock()
		}
	}
}

// killBranch recursively cancels a branch and removes nodes
func (bot *PersistentMinimaxBot) killBranch(node *SearchNode) {
	if node == nil {
		return
	}

	// Cancel node's goroutine
	node.cancel()

	// Recursively kill children
	for _, child := range node.Children {
		bot.killBranch(child)
	}

	// Remove from tree
	delete(bot.tree.nodes, node.ID)
}

// updateDepths recursively updates depths after root change
func (bot *PersistentMinimaxBot) updateDepths(node *SearchNode, newDepth int) {
	if node == nil {
		return
	}

	node.mutex.Lock()
	node.Depth = newDepth
	node.mutex.Unlock()

	for _, child := range node.Children {
		bot.updateDepths(child, newDepth+1)
	}
}

// cleanup shuts down the entire search tree
func (bot *PersistentMinimaxBot) cleanup() {
	if bot.tree != nil {
		bot.tree.cancel()
		bot.tree.wg.Wait()
	}

	bot.rootNode = nil

	// Reinitialize tree
	ctx, cancel := context.WithCancel(context.Background())
	bot.tree = &SearchTree{
		maxDepth:    bot.InitialDepth,
		nodes:       make(map[string]*SearchNode),
		expandQueue: make(chan *SearchNode, 100),
		ctx:         ctx,
		cancel:      cancel,
	}

	go bot.tree.backgroundExpander()
}

// getName implements BotInterface
func (bot *PersistentMinimaxBot) getName() string {
	return bot.Name
}

// getSymbol implements BotInterface
func (bot *PersistentMinimaxBot) getSymbol() byte {
	return bot.Symbol
}

// Close shuts down the bot and cleans up resources
func (bot *PersistentMinimaxBot) Close() {
	bot.cleanup()
}
