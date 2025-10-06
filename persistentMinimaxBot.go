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

	// Initialize search tree with shallower initial depth
	ctx, cancel := context.WithCancel(context.Background())
	bot.tree = &SearchTree{
		maxDepth:    2, // Start shallow and expand gradually
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

	// For now, use a simple approach - just get a valid move quickly
	validMoves := board.GetValidMoves()
	bestMove := ""
	if len(validMoves) > 0 {
		bestMove = validMoves[0] // Take first valid move for now
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
	
	// Find the child corresponding to the move
	newRoot, exists := bot.rootNode.Children[move]
	if !exists {
		// Move not in our search tree, need to cleanup but avoid deadlock
		bot.tree.mutex.Unlock()
		bot.cleanup() // Release lock before cleanup to avoid deadlock
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
	
	bot.tree.mutex.Unlock() // Don't forget to unlock at the end
}

// expandNode runs as a goroutine to expand a search node
func (bot *PersistentMinimaxBot) expandNode(node *SearchNode) {
	bot.tree.wg.Add(1)
	defer bot.tree.wg.Done()

	defer func() {
		// Ensure goroutine signals completion
		select {
		case <-node.goroutine:
		default:
			close(node.goroutine) // Signal that goroutine is running
		}
	}()

	for {
		select {
		case <-node.ctx.Done():
			return // Node was cancelled

		default:
			node.mutex.Lock()

			// Check if we should expand (are we at current max depth or too deep?)
			bot.tree.mutex.RLock()
			currentMaxDepth := bot.tree.maxDepth
			bot.tree.mutex.RUnlock()

			if node.Depth >= currentMaxDepth || node.Depth >= 6 { // Hard limit at depth 6 to prevent explosion
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

				// Limit the number of children to prevent goroutine explosion
				maxChildren := 8
				if len(validMoves) > maxChildren {
					validMoves = validMoves[:maxChildren]
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

					// Safely add to tree nodes map with proper synchronization
					bot.tree.mutex.Lock()
					bot.tree.nodes[childID] = child
					bot.tree.mutex.Unlock()

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
		// Collect child scores first to avoid holding multiple locks
		var childScores []int
		current.mutex.RLock()
		childCount := len(current.Children)
		if childCount > 0 {
			childScores = make([]int, 0, childCount)
			for _, child := range current.Children {
				child.mutex.RLock()
				childScores = append(childScores, child.Score)
				child.mutex.RUnlock()
			}
		}
		isMaximizing := current.IsMaximizing
		current.mutex.RUnlock()

		// Calculate best score without holding locks
		if len(childScores) > 0 {
			bestScore := MIN_INT
			if !isMaximizing {
				bestScore = MAX_INT
			}

			for _, score := range childScores {
				if isMaximizing && score > bestScore {
					bestScore = score
				} else if !isMaximizing && score < bestScore {
					bestScore = score
				}
			}

			// Update score with minimal lock time
			current.mutex.Lock()
			current.Score = bestScore
			current.mutex.Unlock()
		}

		current = current.Parent
	}
}

// backgroundExpander runs background expansion of leaf nodes
func (tree *SearchTree) backgroundExpander() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tree.ctx.Done():
			return

		case <-ticker.C:
			// Gradually increase search depth, but cap it to prevent explosion
			tree.mutex.Lock()
			if tree.maxDepth < 6 { // Cap at depth 6
				tree.maxDepth++
			}
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

	// Get children safely before recursion
	node.mutex.RLock()
	children := make([]*SearchNode, 0, len(node.Children))
	for _, child := range node.Children {
		children = append(children, child)
	}
	node.mutex.RUnlock()

	// Recursively kill children
	for _, child := range children {
		bot.killBranch(child)
	}

	// Remove from tree with proper synchronization
	bot.tree.mutex.Lock()
	delete(bot.tree.nodes, node.ID)
	bot.tree.mutex.Unlock()
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
