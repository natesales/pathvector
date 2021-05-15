import os

import CliParser
import BasicCli
import TerminalUtil


class WireframeCmd(object):
    syntax = """wireframe [ <arg> ]"""
    wireframeArgsRule = CliParser.StringRule("ARG", "Executable arguments and/or switches", 'args')
    data = {'wireframe': 'Run wireframe', '<arg>': wireframeArgsRule}

    # noinspection PyMethodMayBeStatic
    def wireframe(self, mode, args):
        with TerminalUtil.NoTerminalSettings(mode.session_.terminalCtx_):
            command = "sudo wireframe"
            if args:
                command += " " + args
            os.system(command)

    def handler(self, mode, args):
        self.wireframe(mode, args.get("<arg>", None))


# Register the command by adding a new class
BasicCli.EnableMode.addCommandClass(WireframeCmd)
