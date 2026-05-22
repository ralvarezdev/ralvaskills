# ROS2 Recipes

Reference scaffolds extracted from [SKILL.md](SKILL.md). Use these as templates when bootstrapping a workspace or a new package; the rules and *why* stay in the skill body.

## `pixi.toml` — workspace manifest

Replace `<distro>` with `jazzy`, `kilted`, or `lyrical` consistently across the file. Pick once per workspace per [STACK.md](STACK.md#supported-distributions).

```toml
# pixi.toml
[project]
name = "my_robot_ws"
channels = ["https://prefix.dev/conda-forge", "https://prefix.dev/robostack-staging"]
platforms = ["linux-64", "linux-aarch64", "osx-arm64", "win-64"]
description = "My robot workspace"

[dependencies]
python                 = "3.12.*"
# Replace <distro> with `jazzy`, `kilted`, or `lyrical` (consistent across the file)
ros-<distro>-desktop     = "*"
ros-<distro>-rclcpp      = "*"
ros-<distro>-rclpy       = "*"
ros-<distro>-launch      = "*"
ros-<distro>-launch-ros  = "*"
colcon-common-extensions = "*"
compilers              = "*"      # C++ toolchain (gcc / clang / MSVC)
cmake                  = "*"

[tasks]
build = "colcon build --symlink-install"
test  = "colcon test && colcon test-result --verbose"
clean = "rm -rf build/ install/ log/"
run   = { cmd = "ros2 launch my_robot_bringup robot.launch.py", depends-on = ["build"] }
```

```bash
pixi install                # one-time: install ROS + toolchain into .pixi/
pixi shell                  # enter the env
pixi run build              # build everything
pixi run run                # launch the robot
```

## `package.xml` — `ament_cmake` (C++ and `_msgs`)

```xml
<package format="3">
  <name>my_robot_drivers</name>
  <version>0.1.0</version>
  <description>Hardware drivers for My Robot</description>
  <maintainer email="me@example.com">Me</maintainer>
  <license>Apache-2.0</license>

  <buildtool_depend>ament_cmake</buildtool_depend>
  <depend>rclcpp</depend>
  <depend>std_msgs</depend>
  <depend>my_robot_msgs</depend>

  <test_depend>ament_lint_auto</test_depend>
  <test_depend>ament_lint_common</test_depend>

  <export><build_type>ament_cmake</build_type></export>
</package>
```

## `package.xml` — `ament_python`

```xml
<package format="3">
  <name>my_robot_navigation</name>
  <version>0.1.0</version>
  <description>Navigation logic for My Robot</description>
  <maintainer email="me@example.com">Me</maintainer>
  <license>Apache-2.0</license>

  <exec_depend>rclpy</exec_depend>
  <exec_depend>my_robot_msgs</exec_depend>

  <test_depend>ament_copyright</test_depend>
  <test_depend>ament_flake8</test_depend>
  <test_depend>ament_pep257</test_depend>
  <test_depend>python3-pytest</test_depend>

  <export><build_type>ament_python</build_type></export>
</package>
```

## Launch file — Python DSL

Lifecycle node + standard node, with a launch argument, parameter dict, and YAML config. One launch, many configurations.

```python
# launch/robot.launch.py
from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument
from launch.substitutions import LaunchConfiguration
from launch_ros.actions import Node, LifecycleNode

def generate_launch_description():
    use_sim = LaunchConfiguration('use_sim')

    return LaunchDescription([
        DeclareLaunchArgument('use_sim', default_value='false'),

        LifecycleNode(
            package='my_robot_drivers',
            executable='camera_driver',
            name='camera_driver',
            namespace='sensors',
            parameters=[{'frame_rate': 30}],
            condition=...,
        ),

        Node(
            package='my_robot_navigation',
            executable='path_planner',
            name='path_planner',
            output='screen',
            parameters=['config/path_planner.yaml'],
        ),
    ])
```
