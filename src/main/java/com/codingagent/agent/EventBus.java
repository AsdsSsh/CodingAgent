package com.codingagent.agent;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArrayList;
import java.util.function.Consumer;

/**
 * 轻量级进程内发布/订阅事件总线，用于传递 Agent 事件。
 *
 * <p>CLI 层订阅此类事件以实现实时状态渲染（Thought → Tool Call → Observation），
 * 无需将 UI 与 ReAct 循环耦合。订阅者按事件类型注册；
 * 事件在调用线程上同步分发。</p>
 *
 * <p>线程安全：订阅者存储在 {@link CopyOnWriteArrayList} 中；
 * emit 遍历快照——并发 subscribe/emit 安全。</p>
 */
public class EventBus {

    private final Map<Class<?>, List<Consumer<Object>>> subscribers = new ConcurrentHashMap<>();

    /** 向所有注册了该事件精确类型的订阅者发布事件。 */
    @SuppressWarnings("unchecked")
    public <T> void emit(T event) {
        List<Consumer<Object>> handlers = subscribers.get(event.getClass());
        if (handlers != null) {
            for (Consumer<Object> handler : handlers) {
                handler.accept(event);
            }
        }
    }

    /** 为特定类型的事件注册处理器。 */
    @SuppressWarnings("unchecked")
    public <T> void subscribe(Class<T> eventType, Consumer<T> handler) {
        subscribers.computeIfAbsent(eventType, k -> new CopyOnWriteArrayList<>())
            .add((Consumer<Object>) handler);
    }
}
