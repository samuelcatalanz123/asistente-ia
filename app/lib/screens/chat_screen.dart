import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';
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

  static const _clave = 'mensajes-guardados';

  static const _preguntas = [
    'Escríbeme un ejemplo de código en Go 💻',
    'Dame una idea para un proyecto',
    'Cuéntame un chiste de programadores 😄',
    '¿Cómo empiezo a aprender a programar?',
  ];

  @override
  void initState() {
    super.initState();
    _cargar();
  }

  // ---------- #2 Recordar la conversación al cerrar ----------
  Future<void> _cargar() async {
    final prefs = await SharedPreferences.getInstance();
    final guardado = prefs.getString(_clave);
    if (guardado == null) return;
    try {
      final lista = (jsonDecode(guardado) as List)
          .map((e) => Message.fromJson(e as Map<String, dynamic>))
          .toList();
      if (mounted) setState(() => _messages..clear()..addAll(lista));
    } catch (_) {}
  }

  Future<void> _guardar() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(
        _clave, jsonEncode(_messages.map((m) => m.toJson()).toList()));
  }

  void _bajarDelTodo() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scroll.hasClients) {
        _scroll.animateTo(
          _scroll.position.maxScrollExtent,
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        );
      }
    });
  }

  // ---------- #4 Streaming: la respuesta llega trozo a trozo ----------
  Future<void> _send(String text) async {
    text = text.trim();
    if (text.isEmpty || _loading) return;

    setState(() {
      _messages.add(Message(role: 'user', content: text));
      _loading = true;
      _controller.clear();
    });
    _bajarDelTodo();

    final historia = List<Message>.from(_messages);
    final idx = _messages.length; // posición de la respuesta del asistente
    setState(() => _messages.add(const Message(role: 'assistant', content: '…')));

    try {
      final buffer = StringBuffer();
      await for (final chunk in widget.service.streamMessages(historia)) {
        buffer.write(chunk);
        if (!mounted) return;
        setState(() =>
            _messages[idx] = Message(role: 'assistant', content: buffer.toString()));
        _bajarDelTodo();
      }
    } catch (e) {
      if (mounted) {
        setState(() => _messages[idx] = const Message(
            role: 'assistant',
            content: '⚠️ No me pude conectar, inténtalo de nuevo'));
      }
    } finally {
      if (mounted) setState(() => _loading = false);
      _guardar();
      _bajarDelTodo();
    }
  }

  void _nuevaConversacion() async {
    setState(() {
      _messages.clear();
      _loading = false;
    });
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_clave);
  }

  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;

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
            child:
                _messages.isEmpty ? _bienvenida(colors) : _listaMensajes(colors),
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

  Widget _bienvenida(ColorScheme colors) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
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
                .map((p) => ActionChip(label: Text(p), onPressed: () => _send(p)))
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
                maxWidth: MediaQuery.of(context).size.width * 0.8),
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
            child: m.isUser
                ? Text(m.content,
                    style: const TextStyle(color: Colors.white, height: 1.4))
                : _contenidoAsistente(m.content, colors),
          ),
        );
      },
    );
  }

  // ---------- #1 Código bonito: formatea la respuesta del asistente ----------
  Widget _contenidoAsistente(String texto, ColorScheme colors) {
    final partes = texto.split('```');
    final hijos = <Widget>[];
    for (var i = 0; i < partes.length; i++) {
      final parte = partes[i];
      if (i.isOdd) {
        // Bloque de código. Quitamos el nombre del lenguaje de la 1ª línea.
        var codigo = parte;
        final salto = codigo.indexOf('\n');
        if (salto != -1) {
          final primera = codigo.substring(0, salto).trim();
          if (!primera.contains(' ') && primera.length < 15) {
            codigo = codigo.substring(salto + 1);
          }
        }
        hijos.add(_bloqueCodigo(codigo.trimRight()));
      } else if (parte.trim().isNotEmpty) {
        hijos.add(Text.rich(
          TextSpan(children: _spans(parte.trim(), colors)),
          style: TextStyle(color: colors.onSurface, height: 1.4),
        ));
      }
    }
    if (hijos.isEmpty) {
      return Text(texto,
          style: TextStyle(color: colors.onSurface, height: 1.4));
    }
    return Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: hijos);
  }

  // Negrita **...** y código en línea `...`.
  List<InlineSpan> _spans(String text, ColorScheme colors) {
    final spans = <InlineSpan>[];
    final regex = RegExp(r'\*\*(.+?)\*\*|`(.+?)`');
    int last = 0;
    for (final m in regex.allMatches(text)) {
      if (m.start > last) spans.add(TextSpan(text: text.substring(last, m.start)));
      if (m.group(1) != null) {
        spans.add(TextSpan(
            text: m.group(1),
            style: const TextStyle(fontWeight: FontWeight.bold)));
      } else if (m.group(2) != null) {
        spans.add(TextSpan(
            text: m.group(2),
            style: TextStyle(
                fontFamily: 'monospace',
                backgroundColor: colors.surfaceContainerHighest)));
      }
      last = m.end;
    }
    if (last < text.length) spans.add(TextSpan(text: text.substring(last)));
    return spans;
  }

  Widget _bloqueCodigo(String codigo) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(vertical: 6),
      decoration: BoxDecoration(
        color: const Color(0xFF1E1E2E),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Stack(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 34, 14, 14),
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: SelectableText(
                codigo,
                style: const TextStyle(
                    fontFamily: 'monospace',
                    color: Color(0xFFE7E7F0),
                    fontSize: 13.5),
              ),
            ),
          ),
          Positioned(
            top: 2,
            right: 2,
            child: TextButton.icon(
              onPressed: () {
                Clipboard.setData(ClipboardData(text: codigo));
                ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
                    content: Text('¡Código copiado!'),
                    duration: Duration(seconds: 1)));
              },
              icon: const Icon(Icons.copy, size: 14, color: Colors.white70),
              label: const Text('Copiar',
                  style: TextStyle(color: Colors.white70, fontSize: 12)),
            ),
          ),
        ],
      ),
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
