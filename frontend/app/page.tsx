'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';

interface Task {
  ID: number;
  title: string;
  status: string;
}

export default function Home() {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [newTask, setNewTask] = useState('');
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  // 環境変数からAPIのURLを取得
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  useEffect(() => {
    // 1. トークンチェック（なければログイン画面へ強制送還）
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/login');
      return;
    }
    fetchTasks();
  }, [router]);

  // 認証付きでタスクを取得する関数
  const fetchTasks = async () => {
    const token = localStorage.getItem('token');
    try {
      const res = await fetch(`${apiUrl}/tasks`, {
        headers: {
          'Authorization': `Bearer ${token}` // 【重要】ここに鍵を載せる
        }
      });

      if (res.status === 401) {
        // トークンが期限切れならログアウトさせる
        handleLogout();
        return;
      }

      const data = await res.json();
      setTasks(data);
    } catch (error) {
      console.error('Failed to fetch tasks:', error);
    } finally {
      setLoading(false);
    }
  };

  const addTask = async () => {
    if (!newTask) return;
    const token = localStorage.getItem('token');

    await fetch(`${apiUrl}/tasks`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}` // 【重要】ここにも鍵を載せる
      },
      body: JSON.stringify({ title: newTask, status: 'Pending' }),
    });
    setNewTask('');
    fetchTasks();
  };

  const deleteTask = async (id: number) => {
    const token = localStorage.getItem('token');
    await fetch(`${apiUrl}/tasks/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${token}` // 【重要】ここにも
      }
    });
    fetchTasks();
  };

  const handleLogout = () => {
    localStorage.removeItem('token'); // 鍵を捨てる
    router.push('/login'); // ログイン画面へ戻る
  };

  // ロード中は何も表示しない（チラつき防止）
  if (loading) return <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">Loading...</div>;

  return (
    <div className="min-h-screen bg-gray-900 text-white font-sans">
      <div className="max-w-xl mx-auto p-8">
        <header className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
            My Tasks
          </h1>
          <button 
            onClick={handleLogout}
            className="text-sm text-gray-400 hover:text-white border border-gray-600 px-3 py-1 rounded"
          >
            Logout
          </button>
        </header>

        <div className="flex gap-2 mb-8">
          <input
            type="text"
            className="flex-1 p-3 rounded-lg bg-gray-800 border border-gray-700 text-white focus:outline-none focus:border-blue-500 transition"
            placeholder="What needs to be done?"
            value={newTask}
            onChange={(e) => setNewTask(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && addTask()}
          />
          <button
            onClick={addTask}
            className="bg-blue-600 hover:bg-blue-500 text-white px-6 py-3 rounded-lg font-bold transition shadow-lg hover:shadow-blue-500/20"
          >
            Add
          </button>
        </div>

        <ul className="space-y-3">
          {tasks.map((task) => (
            <li
              key={task.ID}
              className="flex justify-between items-center bg-gray-800 p-4 rounded-lg shadow border border-gray-700 hover:border-gray-600 transition group"
            >
              <span className="text-lg">{task.title}</span>
              <button
                onClick={() => deleteTask(task.ID)}
                className="text-gray-500 hover:text-red-400 opacity-0 group-hover:opacity-100 transition"
              >
                Delete
              </button>
            </li>
          ))}
        </ul>
        
        {tasks.length === 0 && (
          <p className="text-center text-gray-500 mt-10">No tasks yet. Add one above!</p>
        )}
      </div>
    </div>
  );
}