"use client";

import { useState, useEffect } from "react";

// タスクの型定義（Goの構造体と合わせる）
type Task = {
  ID: number;
  title: string;
  status: string;
};

export default function Home() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTaskTitle, setNewTaskTitle] = useState("");

  // 1. マウント時にタスク一覧を取得
  useEffect(() => {
    fetchTasks();
  }, []);

  const fetchTasks = async () => {
    try {
      const res = await fetch("http://localhost:8080/tasks");
      const data = await res.json();
      setTasks(data || []); // nullの場合は空配列にする
    } catch (err) {
      console.error("Fetch error:", err);
    }
  };

  // 2. タスク追加処理
  const addTask = async () => {
    if (!newTaskTitle) return;
    try {
      await fetch("http://localhost:8080/tasks", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ title: newTaskTitle, status: "pending" }),
      });
      setNewTaskTitle(""); // 入力欄をクリア
      fetchTasks(); // リストを再取得
    } catch (err) {
      console.error("Add error:", err);
    }
  };

  // 3. タスク削除処理
  const deleteTask = async (id: number) => {
    try {
      await fetch(`http://localhost:8080/tasks/${id}`, {
        method: "DELETE",
      });
      fetchTasks(); // リストを再取得
    } catch (err) {
      console.error("Delete error:", err);
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-2xl mx-auto bg-white rounded-lg shadow-md p-6">
        <h1 className="text-2xl font-bold mb-6 text-gray-800">タスク管理アプリ</h1>

        {/* 入力フォーム */}
        <div className="flex gap-2 mb-8">
          <input
            type="text"
            className="flex-1 border border-gray-300 rounded px-4 py-2 text-gray-800 focus:outline-none focus:border-blue-500"
            placeholder="新しいタスクを入力..."
            value={newTaskTitle}
            onChange={(e) => setNewTaskTitle(e.target.value)}
          />
          <button
            onClick={addTask}
            className="bg-blue-600 hover:bg-blue-700 text-white font-bold px-6 py-2 rounded transition"
          >
            追加
          </button>
        </div>

        {/* タスクリスト */}
        <ul className="space-y-3">
          {tasks.map((task) => (
            <li
              key={task.ID}
              className="flex items-center justify-between p-4 bg-gray-50 border rounded hover:bg-gray-100 transition"
            >
              <span className="text-gray-700">{task.title}</span>
              <button
                onClick={() => deleteTask(task.ID)}
                className="text-red-500 hover:text-red-700 text-sm font-semibold"
              >
                削除
              </button>
            </li>
          ))}
          {tasks.length === 0 && (
            <p className="text-center text-gray-400 py-4">タスクがありません</p>
          )}
        </ul>
      </div>
    </div>
  );
}