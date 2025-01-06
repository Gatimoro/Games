import pygame
import sys
import random
from collections import deque
# Initialize Pygame

# Constants
WIDTH, HEIGHT = 1080, 800  
CELL_SIZE = 40           
GRID_COLOR = (200, 200, 200) 
BG_COLOR = (0, 200, 0)     
SNAKE_COLOR = (0, 255, 255)    
FOOD_COLOR = (255, 0, 0)     
WALL_COLOR=(47,32,32)
# Create the screen

clock = pygame.time.Clock()
# Draw the grid with updated cell colors
def draw_grid(snake, wall, food):
    for x in range(0, WIDTH, CELL_SIZE):
        for y in range(0, HEIGHT, CELL_SIZE):
            grid_x = x // CELL_SIZE
            grid_y = y // CELL_SIZE
            if (grid_x, grid_y) in snake:
                color = SNAKE_COLOR  
            elif (grid_x, grid_y) in wall:
                color = WALL_COLOR
            elif (grid_x, grid_y) in food:
                color = FOOD_COLOR
            else:
                color = BG_COLOR 
            
            rect = pygame.Rect(x, y, CELL_SIZE, CELL_SIZE)
            pygame.draw.rect(screen, color, rect)
            pygame.draw.rect(screen, GRID_COLOR, rect, 1)  

# Game loop
width_cells=WIDTH//CELL_SIZE
height_cells=HEIGHT//CELL_SIZE
playing=True
while playing:
    pygame.init()
    screen = pygame.display.set_mode((WIDTH, HEIGHT))
    
    head=(5,5)
    snake={head}
    snakeq=deque([head])
    food={(random.randint(0, width_cells-1),random.randint(0, height_cells-1))}
    wall=set()
    score=0
    speed=6
    pygame.display.set_caption('DaSnake')
    direction=[1,0]
    while True:
        for event in pygame.event.get():
            if event.type == pygame.KEYDOWN:
                if last:
                    if event.key == pygame.K_UP:
                        direction=(0,-1)
                    elif event.key == pygame.K_DOWN:
                        direction=(0,1)
                else:                    
                    if event.key == pygame.K_LEFT:
                        direction=(-1,0)
                    elif event.key == pygame.K_RIGHT:
                        direction=(1,0)
            if event.type == pygame.QUIT:
                pygame.quit()
                sys.exit()
        #update new head position
        last=direction[0]
        head=((head[0] + direction[0])%width_cells , (head[1] + direction[1])%height_cells)
        #calculate effect of head
        if head in food:
            food.remove(head)
            score+=1
            pygame.display.set_caption(f"Score: {score}")
            if not score%5:
                speed+=1
            if not score%10:
                appledrop=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
                while appledrop==head or appledrop in food or appledrop in snake or appledrop in wall or appledrop == (head[0]+direction[0],head[1] + direction[1]):
                    appledrop=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
                food.add(appledrop)
            wal=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
            while wal==head or wal in food or wal in snake or wal in wall or 2>= abs( head[0]+2*direction[0]+head[1] + 2*direction[1] -wal[0]-wal[1]):
                wal=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
            wall.add(wal)
            appledrop=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
            while appledrop == head or appledrop in food or appledrop in snake or appledrop in wall or appledrop == (head[0]+direction[0],head[1] + direction[1]):
                appledrop=(random.randint(0,width_cells-1),random.randint(0,height_cells-1))
            food.add(appledrop)
        
        else:
            snake.remove(snakeq.popleft())
        if (head in snake or head in wall) :
            print(f'Your score was {score}!')
            break
        snake.add(head)
        
        snakeq.append(head)
        screen.fill(BG_COLOR) 
        draw_grid(snake,wall,food)  
        pygame.display.flip()
        
        clock.tick(speed)
    
