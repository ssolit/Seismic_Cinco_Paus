#var currPos = [0, 0]
#
#
##func _physics_process(delta):
	##move()
	##
##func move():
	##input_movement = Input.get_vector("left", "right", "up", "down")
	##
	##if input_movement != Vector2.ZERO:
		##velocity = input_movement * speed
	##
	##if input_movement == Vector2.ZERO:
		##velocity = Vector2.ZERO
		##
	##move_and_slide()
#
#func _input(event):
	#if event.is_action_pressed("right"):
		#if currPos[0] <= 96:
			#currPos[0] += 32
	#elif event.is_action_pressed("left"):
		#if currPos[0] >= 32:
			#currPos[0] -= 32	
	#elif event.is_action_pressed("up"):
		#if currPos[1] >= 32:
			#currPos[1] -= 32
	#elif event.is_action_pressed("down"):
		#if currPos[1] <= 96:	
			#currPos[1] += 32
#
	#self.position = Vector2(currPos[0], currPos[1])
	
extends Area2D



const MAX_HEALTH = 5
var health = MAX_HEALTH
var previous_move

var animation_speed = 4

@export var moving = false
var tile_size = 64
var inputs = {
	"right": Vector2.RIGHT,
	"left": Vector2.LEFT,
	"up": Vector2.UP,
	"down": Vector2.DOWN
}

@onready var ray = $RayCast2D
@onready var wand_animations = $"../WandAnimations"
	
func _ready():
	$"../LifeBar/Life1".play("hearts")
	$"../LifeBar/Life2".play("hearts")
	$"../LifeBar/Life3".play("hearts")
	$"../LifeBar/Life4".play("hearts")
	$"../LifeBar/Life5".play("hearts")
	update_health_ui()
	position = position.snapped(Vector2.ONE * tile_size)
	#position += Vector2.ONE * tile_size / 


func _process(delta):
	$Sprite.play("idle")


func update_health_ui():
	for i in range(MAX_HEALTH):
		$"../LifeBar".get_child(i).visible = health > i


func _input(event: InputEvent) -> void:
	if event.is_action_pressed("ui_accept"):
		#damage()
		wand_animations.position = Vector2(20, -23)
		wand_animations.play("lightning_strike")
	


func damage(damage) -> void:
	health -= damage
	if health == 0:
		self.visible = false
		
	update_health_ui()


func _unhandled_input(event):
	if moving:
		return
	for dir in inputs.keys():
		if event.is_action_pressed(dir):
			move(dir)


func move(dir):
	ray.target_position = inputs[dir] * tile_size
	ray.force_raycast_update()
	if !ray.is_colliding():
		previous_move = dir 
		#position += inputs[dir] * tile_size
		var tween = get_tree().create_tween()
		tween.tween_property(self, "position", position + inputs[dir] * tile_size, 1.0/animation_speed).set_trans(Tween.TRANS_SINE)
		moving = true 
		await tween.finished
		moving = false
	else:
		$AnimationPlayer.play("hit_wall")


func _on_area_entered(area):
	print("do stuff")
	if (area.name == "Enemy - 2 damage" or area.name == "Enemy - 1 damage") && moving == true:
		area.damage()
		match area.previous_move:
			"right": area.move("left")
			"left": area.move("right")
			"up": area.move("down")
			"down": area.move("up")
		
		
		

