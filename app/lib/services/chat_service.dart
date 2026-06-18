import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/message.dart';

class ChatService {
  final String baseUrl;
  final http.Client client;

  ChatService({required this.baseUrl, http.Client? client})
      : client = client ?? http.Client();

  /// Envía la conversación y devuelve la respuesta completa de una vez.
  /// (Se mantiene como respaldo y para las pruebas.)
  Future<String> sendMessages(List<Message> messages) async {
    final response = await client.post(
      Uri.parse('$baseUrl/chat'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'messages': messages.map((m) => m.toJson()).toList()}),
    );

    if (response.statusCode != 200) {
      throw Exception('No me pude conectar, inténtalo de nuevo');
    }
    final data = jsonDecode(response.body) as Map<String, dynamic>;
    return data['reply'] as String;
  }

  /// Envía la conversación y va devolviendo la respuesta trozo a trozo
  /// (streaming con Server-Sent Events), igual que la web.
  Stream<String> streamMessages(List<Message> messages,
      {String modo = 'amigable'}) async* {
    final request = http.Request('POST', Uri.parse('$baseUrl/chat/stream'));
    request.headers['Content-Type'] = 'application/json';
    request.body = jsonEncode({
      'messages': messages.map((m) => m.toJson()).toList(),
      'modo': modo,
    });

    final response = await client.send(request);
    if (response.statusCode != 200) {
      throw Exception('No me pude conectar, inténtalo de nuevo');
    }

    final lineas =
        response.stream.transform(utf8.decoder).transform(const LineSplitter());
    await for (final linea in lineas) {
      if (!linea.startsWith('data: ')) continue;
      final data = linea.substring(6);
      if (data == '[DONE]') break;
      Map<String, dynamic>? obj;
      try {
        obj = jsonDecode(data) as Map<String, dynamic>;
      } catch (_) {
        continue;
      }
      if (obj['error'] != null) {
        throw Exception(obj['error']);
      }
      final t = obj['t'];
      if (t is String) yield t;
    }
  }
}
