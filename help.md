### 运行总体流程

1. ui.Run()
2. director.Start()
3. 先使用 setView() 进行初始化操作，之后使用 director.Run() 开始正式运行
4. director.Step() 单步执行
5. view.Update()
    1. 绑定纹理
    2. 创建纹理
    3. 渲染纹理到画布上

在游戏界面会多出来一个Console属性，该属性用于存储NES的相关属性信息，事实上Console就可以作为NES游戏ROM属性的代表。

光栅化：向量（几何图形） -> 位图（像素图形）

<https://github.com/fogleman/nes>

使用到的跨平台声音和图像库：
* 声音库：portaudio
* 图形库：opengl


### 核心技术

1. opengl的图像渲染
2. 游戏rom文件的解析
3. nes游戏机的模拟

显示出来的图像由**背景**和**活动块**共同组成。
