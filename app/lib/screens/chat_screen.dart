import 'package:flutter/material.dart';
import '../models/message.dart';
import '../services/chat_service.dart';

class ChatScreen extends StatefulWidget {
  final ChatService service;
  final VoidCallback onToggleTheme;
  final bool isDark;

  const ChatScreen({
    super.key,
    required this.service,
    required this.onToggleTheme,
    required this.isDark,
  });

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _controller = TextEditingController();
  final _scroll = ScrollController();
  final _messages = <Message>[];
  bool _loading = false;

  // Preguntas de ejemplo que se muestran al empezar.
  static const _preguntas = [
    'Escríbeme un ejemplo de código en Go 💻',
    'Dame una idea para un proyecto',
    'Cuéntame un chiste de programadores 😄',
    '¿Cómo empiezo a aprender a programar?',
  ];

  void _bajarDelTodo() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.animateTo(
          _scroll.position.maxScrollExtent,
          duration: const Duration(milliseconds: 250),
          curve: Curves.easeOut,
        );
      }
    });
  }

  Future<void> _send(String text) async {
    text = text.trim();
    if (text.isEmpty || _loading) return;

    setState(() {
      _messages.add(Message(role: 'user', content: text));
      _loading = true;
      _controller.clear();
    });
    _bajarDelTodo();

    try {
      final reply = await widget.service.sendMessages(_messages);
      if (!mounted) return;
      setState(() => _messages.add(Message(role: 'assistant', content: reply)));
    } catch (e) {
      if (!mounted) return;
      setState(() => _messages.add(Message(
          role: 'assistant',
          content: '⚠️ No me pude conectar, inténtalo de nuevo')));
    } finally {
      if (mounted) setState(() => _loading = false);
      _bajarDelTodo();
    }
  }

  void _nuevaConversacion() {
    setState(() {
      _messages.clear();
      _loading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colors = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        backgroundColor: colors.primary,
        foregroundColor: Colors.white,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: const [
            Text('Asistente IA',
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 18)),
            Text('Creado por Samuel',
                style: TextStyle(fontSize: 12, color: Colors.white70)),
          ],
        ),
        actions: [
          IconButton(
            tooltip: 'Modo claro/oscuro',
            icon: Icon(widget.isDark ? Icons.light_mode : Icons.dark_mode),
            onPressed: widget.onToggleTheme,
          ),
          IconButton(
            tooltip: 'Nueva conversación',
            icon: const Icon(Icons.refresh),
            onPressed: _nuevaConversacion,
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _messages.isEmpty ? _bienvenida(colors) : _listaMensajes(colors),
          ),
          if (_loading)
            const Padding(
              padding: EdgeInsets.all(10),
              child: Row(
                children: [
                  SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2)),
                  SizedBox(width: 10),
                  Text('escribiendo…'),
                ],
              ),
            ),
          _barraEscribir(colors),
        ],
      ),
    );
  }

  // Pantalla de bienvenida con las preguntas de ejemplo.
  Widget _bienvenida(ColorScheme colors) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          const SizedBox(height: 20),
          Icon(Icons.smart_toy_outlined, size: 64, color: colors.primary),
          const SizedBox(height: 12),
          const Text('¡Hola! Soy tu asistente.',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600)),
          const SizedBox(height: 4),
          const Text('Prueba con una de estas:',
              style: TextStyle(color: Colors.grey)),
          const SizedBox(height: 16),
          Wrap(
            alignment: WrapAlignment.center,
            spacing: 8,
            runSpacing: 8,
            children: _preguntas
                .map((p) => ActionChip(
                      label: Text(p),
                      onPressed: () => _send(p),
                    ))
                .toList(),
          ),
        ],
      ),
    );
  }

  Widget _listaMensajes(ColorScheme colors) {
    return ListView.builder(
      controller: _scroll,
      padding: const EdgeInsets.all(14),
      itemCount: _messages.length,
      itemBuilder: (context, i) {
        final m = _messages[i];
        return Align(
          alignment: m.isUser ? Alignment.centerRight : Alignment.centerLeft,
          child: Container(
            margin: const EdgeInsets.symmetric(vertical: 5),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 11),
            constraints: BoxConstraints(
                maxWidth: MediaQuery.of(context).size.width * 0.78),
            decoration: BoxDecoration(
              color: m.isUser ? colors.primary : colors.surfaceContainerHighest,
              borderRadius: BorderRadius.only(
                topLeft: const Radius.circular(18),
                topRight: const Radius.circular(18),
                bottomLeft: Radius.circular(m.isUser ? 18 : 5),
                bottomRight: Radius.circular(m.isUser ? 5 : 18),
              ),
              boxShadow: [
                BoxShadow(
                    color: Colors.black.withValues(alpha: 0.06),
                    blurRadius: 6,
                    offset: const Offset(0, 2)),
              ],
            ),
            child: Text(
              m.content,
              style: TextStyle(
                color: m.isUser ? Colors.white : colors.onSurface,
                height: 1.4,
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _barraEscribir(ColorScheme colors) {
    return SafeArea(
      top: false,
      child: Padding(
        padding: const EdgeInsets.fromLTRB(10, 6, 10, 10),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _controller,
                textInputAction: TextInputAction.send,
                decoration: InputDecoration(
                  hintText: 'Escribe un mensaje…',
                  filled: true,
                  contentPadding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                ),
                onSubmitted: _send,
              ),
            ),
            const SizedBox(width: 8),
            CircleAvatar(
              backgroundColor: colors.primary,
              child: IconButton(
                icon: const Icon(Icons.send, color: Colors.white),
                onPressed: () => _send(_controller.text),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
