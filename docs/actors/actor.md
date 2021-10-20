## 简介

> actor是 [Entity](../entity/entity.md)的运行时模式, 用于维护`Entity`的实时状态和提供`Entity`的具体行为。





## Actor

一个Actor指的是一个最基本的计算单元。 它能接收一个消息并且基于其执行计算。在我们的设计中actor有三部分组成：`mailbox`，`state`，`coroutine`，`mailbox` 是actor的信箱，用于接受外部输入；`state` 是actor自身的状态；`coroutine` 是actor附着的协程。

![actor-consistents](../images/actor.png)


外部消息发送到 `actor` 的 `mailbox`，`actor`从`mailbox` 获取消息，然后actor执行消息并更新自身状态。

![dispatch-message-to-actor](../images/dispatch-msg-to-actor.png)


### 执行流程分析

1. actor 被创建，初始化actor，actor内coroutine运行。
2. 外部消息异步发送到actor的mailbox，mailbox是一个简单的消息队列，mailbox接受到消息触发actor内的coroutine唤醒。
3. 唤醒的actor.coroutine消费mailbox，执行计算并更新自身状态。
4. actor.coroutine处理完mailbox内所有消息后阻塞等待唤醒。
5. 重复2-4过程。



## Reactor


> Don't call us, we'll call you. Reactor模式是一种典型的事件驱动的编程模型，Reactor逆置了程序处理的流程。事件驱动模型是一个状态机，包含了状态(state), 输入事件(input-event), 状态转移(transition), 状态转移即状态到输入事件的一组映射。



![state-marchine-consistents](../images/reactor.png)


`State Marchine` 通过注册的回调来接受消息并更新自身状态。


![dispatch-message-to-reactor](../images/dispatch-msg-to-reactor.png)

### 执行流程分析

1. 创建并初始化状态机。
2. `State Marchine Pool`从`Message Queue`消费到一条事件。
3. 根据事件拿到事件对应的`State Marchine`的上下文。
4. 以事件为输入执行`State Marchine`注册的回调， 完成`State Marchine`的状态更新。
5. 重复执行2-4过程。



## Actor & Reactor

1. 每个actor都有自己的`mailbox`，所以消息的处理不会直接在`Queue`处阻塞，大大的提高高消息的并发处理性能。
2. 每一个`actor`都拥有一个自己的`coroutine`，但是在`iot`场景下各个`actor`的消息频率是不同的，部分`actor`过载，部分`actor`活跃，部分`actor`不活跃，`actor`的负载不同会导致：

    - `OS`对`actor`的工作负载是没有感知的，所以对每一个`actor`中的`coroutine`调度是没有区别的，这样会导致许多的无效调度，`coroutine`负载不均衡，空耗`CPU`资源。
    - 不太活跃的`actor`可能引起`actor`在分布式系统中频繁的调度。
3. reactor使用`coroutine pool`+`State Machine`的模式来实现，`pool`的存在可以使得`coroutine`的调度是`有效`的且`均衡`的，提高`coroutine`的`CPU`时间片的使用效率，提升`CPU`资源的使用效率。
4. reactor消费从Queue消费消息，并使用`coroutine pool`来并行计算，可能存在多个coroutine阻塞在同一个`State Machine`上，从而降低`coroutine`的使用效率。
5. reactor使用`coroutine pool`来顺序消费Queue，相对而言其并行处理效率较之actor更低。










## References






