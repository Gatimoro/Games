from random import choice
symbol = [" ", "X", "O"]
all_moves = [(i, j) for i in range(3) for j in range(3)]

class TICtok:
    
    def __init__(self):
        self.board = [[0,0,0],[0,0,0],[0,0,0]]
        self.lines = [0,0,0,0,0,0,0,0]
        self.next_player = 1 
    def show(self):
        c=0
        print(69 * '\n')
        for row in self.board[::-1]:
            print(f" {symbol[row[0]]} | {symbol[row[1]]} | {symbol[row[2]]}")
            if c != 2:
                print("-"*11)
                c+=1
    def move(self, x, y, reverse = False):
        #if reverse and self.board[x][y] != 0 or (not reverse) and self.board[x][y] == 0:
            #return "Invalid Move"
        
        if reverse:
            if self.board[x][y] == 0:
                return False
            self.board[x][y] = 0
        else:
            if self.board[x][y] != 0:
                print("That cell is taken!")
                return False 
            self.board[x][y] = self.next_player
        self.lines[x] += self.next_player
        self.lines[3+y] += self.next_player
        if x == y:
            self.lines[6] += self.next_player
        if x == 2-y:
            self.lines[7] += self.next_player

        self.next_player *= -1
        return True
    #check if the move we just played loses
    def check4loss(self,x, y):
        self.move(x,y)
        #already lost, return true and viceversa
        if -3 * self.next_player in self.lines:
            self.move(x,y,True)
            return False 
        if 3 * self.next_player in self.lines:
            self.move(x,y, True)
            return True
        
        #check every possible response to the (our computer) move being checked
        for i,j in all_moves:
            if self.board[i][j] != 0:
                continue
            self.move(i,j)
            
            #and if every computer response to said response loses, that's a forced mate. So return true because our original move leads to forced mate. 
            #note that we also check if we've just been mated. Nice. 
            next_moves_lose = [self.check4loss(i,j) for i,j in all_moves if self.board[i][j] == 0]
            if -3 * self.next_player in self.lines or any(next_moves_lose) and all(next_moves_lose):
                self.move(i,j, True)
                self.move(x,y, True)
                return True
            self.move(i,j,True)
        self.move(x,y,True)
        return False 

    def computer_move(self):
        candidates = []
        badmoves = []
        #check computer move
        for x, y in all_moves:
            if self.board[x][y] != 0:
                continue

            if self.check4loss(x,y):
                badmoves.append((x,y))
            else:
                candidates.append((x,y))
        print(candidates, badmoves)
        if not candidates:
            nx, ny = choice(badmoves)
            move(nx, ny)
            return nx, ny, "aaaaaaaaaAAAAAAAAroroggggggggggghh"
        for x,y in candidates:
            self.move(x,y)
            next_moves_lose = [self.check4loss(i,j) for i,j in all_moves if self.board[i][j] == 0]
            if -3 * self.next_player in self.lines or any(next_moves_lose) and all(next_moves_lose):
                return x, y, choice(("Sooooooooo baaaaaaad", "Lost to a toaster lul", "Dumbass!","Get on my level... busssywussy"))
            self.move(x,y,True)
        nx, ny = choice(candidates)
        self.move(nx, ny)
        return nx,ny, None

    def start(self, computerStarts = False):
        self.__init__()
        self.show()
        if computerStarts:
            x, y = choice(all_moves)
            self.next_player *= -1
            self.move(x,y)
            self.show()
            print(f"The computer chose {chr(97+y)}{x+1}")
        while True: 
            while True:
                try:
                    usermove = input("what's your next move?\n").strip()
                    
                    x, y = int(usermove[-1])-1, int(usermove[0]) - 1 if not(97 <= ord(usermove[0])<100) else ord(usermove[0]) - 97
                    
                    if self.move(x,y):
                        break
                except:
                    print('yo mad bruv')
            self.show()
            if self.game_finished(): return 
            if not(any(0 in row for row in self.board)):
                break
            computer_x, computer_y, dialog = self.computer_move()
            self.show()
            print(f"The computer chose {chr(97 + computer_y)}{computer_x + 1}.")
            if dialog: print("He says:",dialog)
            if self.game_finished(): return
        if 3 not in self.lines:
            print("a tie! who would have thought?!")
    def game_finished(self):
        if -3 in self.lines:
            print('BAM BAM BAM BAM BAM\n L BOZO CLOWN BOMBOCLAATTTT')
            return True
        if 3 in self.lines:
            print("WOW!!!\nYOU REALLY DID IT???!!!")
            return True
        if not any(0 in row for row in self.board):
            print("The rare tie, who would've expected it...?")
            return True
        return False 
        
        
        
game = TICtok()
game.start(True)
